package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/snowflake"

	"gorm.io/gorm"
)

type ThemeContentService struct {
	tcRepo  *repository.ThemeContentRepository
	subSvc  *SubscriptionService
}

func NewThemeContentService(tcRepo *repository.ThemeContentRepository, subSvc *SubscriptionService) *ThemeContentService {
	return &ThemeContentService{tcRepo: tcRepo, subSvc: subSvc}
}

type CreateThemeContentRequest struct {
	ID           int64  `json:"id" vd:"$>0"`
	UserID       int64  `json:"user_id" vd:"$>0"`
	ScreenName   string `json:"screen_name"`
	Text         string `json:"text"`
	MetaKeywords string `json:"meta_keywords"`
}

type UpdateThemeContentRequest struct {
	ScreenName   string `json:"screen_name"`
	Text         string `json:"text"`
	MetaKeywords string `json:"meta_keywords"`
}

type TimelineUser struct {
	ScreenName string `json:"screen_name"`
}

type HomeTimeline struct {
	ID           int64        `json:"id"`
	UserID       int64        `json:"user_id"`
	User         TimelineUser `json:"user"`
	Text         string       `json:"text"`
	MetaKeywords string       `json:"meta_keywords"`
	CreatedAt    int64        `json:"created_at"`
	EditedAt     int64        `json:"edited_at"`
}

type ThemeContentBatchRequest struct {
	HomeTimeline []HomeTimeline `json:"home_timeline" vd:"len($)>0"`
}

type SaveTimelineRequest struct {
	UserID       int64                    `json:"user_id" vd:"$>0"`
	HomeTimeline []map[string]interface{} `json:"home_timeline" vd:"len($)>0"`
}

type ThemeContentListResponse struct {
	Data    []model.XqThemeContent `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
}

type ThemeContentSearchResponse struct {
	Data    []model.XqThemeContent `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	Keyword string                 `json:"keyword"`
}

func (s *ThemeContentService) Create(req *CreateThemeContentRequest) (*model.XqThemeContent, error) {
	exists, err := s.tcRepo.Exists(req.ID, req.UserID)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	if exists {
		return nil, errcode.ErrThemeContentDuplicate
	}

	now := time.Now()
	tc := &model.XqThemeContent{
		ID:           req.ID,
		UserID:       req.UserID,
		ScreenName:   req.ScreenName,
		Text:         req.Text,
		MetaKeywords: req.MetaKeywords,
		CreatedAt:    &now,
		EditedAt:     &now,
	}
	if err := s.tcRepo.Create(tc); err != nil {
		return nil, errcode.ErrInternal
	}
	return tc, nil
}

func (s *ThemeContentService) BatchCreate(req *ThemeContentBatchRequest) (int, error) {
	subMap, err := s.subSvc.GetAllEnabledMap()
	if err != nil {
		return 0, errcode.ErrInternal
	}

	var records []model.XqThemeContent
	for _, item := range req.HomeTimeline {
		if _, ok := subMap[item.UserID]; !ok {
			continue
		}

		id := snowflake.NextID()
		tc := model.XqThemeContent{
			ID:           id,
			UserID:       item.UserID,
			ScreenName:   item.User.ScreenName,
			Text:         item.Text,
			MetaKeywords: item.MetaKeywords,
		}
		if item.CreatedAt > 0 {
			t := time.UnixMilli(item.CreatedAt)
			tc.CreatedAt = &t
		}
		if item.EditedAt > 0 {
			t := time.UnixMilli(item.EditedAt)
			tc.EditedAt = &t
		}
		records = append(records, tc)
	}

	if len(records) == 0 {
		return 0, nil
	}

	if err := s.tcRepo.BatchCreateIgnore(records); err != nil {
		return 0, errcode.ErrInternal
	}
	return len(records), nil
}

func (s *ThemeContentService) Update(id, userID int64, req *UpdateThemeContentRequest) (*model.XqThemeContent, error) {
	_, err := s.tcRepo.GetByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrThemeContentNotFound
		}
		return nil, errcode.ErrInternal
	}

	updates := map[string]any{"edited_at": time.Now()}
	if req.ScreenName != "" {
		updates["screen_name"] = req.ScreenName
	}
	if req.Text != "" {
		updates["text"] = req.Text
	}
	if req.MetaKeywords != "" {
		updates["meta_keywords"] = req.MetaKeywords
	}

	if err := s.tcRepo.Update(id, userID, updates); err != nil {
		return nil, errcode.ErrInternal
	}

	tc, err := s.tcRepo.GetByIDAndUserID(id, userID)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	return tc, nil
}

func (s *ThemeContentService) Delete(id, userID int64) error {
	rows, err := s.tcRepo.Delete(id, userID)
	if err != nil {
		return errcode.ErrInternal
	}
	if rows == 0 {
		return errcode.ErrThemeContentNotFound
	}
	return nil
}

func (s *ThemeContentService) GetByID(id, userID int64) (*model.XqThemeContent, error) {
	tc, err := s.tcRepo.GetByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrThemeContentNotFound
		}
		return nil, errcode.ErrInternal
	}
	return tc, nil
}

func (s *ThemeContentService) GetByUserID(userID int64, limit, offset int) (*ThemeContentListResponse, error) {
	list, total, err := s.tcRepo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	return &ThemeContentListResponse{Data: list, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *ThemeContentService) GetAll(limit, offset int) (*ThemeContentListResponse, error) {
	list, total, err := s.tcRepo.GetAll(limit, offset)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	return &ThemeContentListResponse{Data: list, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *ThemeContentService) Exists(id, userID int64) (bool, error) {
	exists, err := s.tcRepo.Exists(id, userID)
	if err != nil {
		return false, errcode.ErrInternal
	}
	return exists, nil
}

func (s *ThemeContentService) Search(keyword string, userID int64, limit, offset int) (*ThemeContentSearchResponse, error) {
	list, total, err := s.tcRepo.Search(keyword, userID, limit, offset)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	return &ThemeContentSearchResponse{
		Data: list, Total: total, Limit: limit, Offset: offset, Keyword: keyword,
	}, nil
}

func (s *ThemeContentService) SaveTimeline(req *SaveTimelineRequest) error {
	exists, err := s.tcRepo.SubscriptionExistsByUserID(req.UserID)
	if err != nil {
		return errcode.ErrInternal
	}
	if !exists {
		return errcode.ErrSubscriptionNotFound
	}

	jsonData, err := json.Marshal(req.HomeTimeline)
	if err != nil {
		return errcode.ErrInternal
	}

	now := time.Now()
	tc := &model.XqThemeContent{
		ID:           req.UserID,
		UserID:       req.UserID,
		ScreenName:   "雪球帖子数据",
		Text:         string(jsonData),
		MetaKeywords: fmt.Sprintf("帖子数量: %d", len(req.HomeTimeline)),
		CreatedAt:    &now,
		EditedAt:     &now,
	}
	return s.tcRepo.Create(tc)
}

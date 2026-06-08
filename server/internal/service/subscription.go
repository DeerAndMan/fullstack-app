package service

import (
	"errors"
	"strconv"
	"strings"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/snowflake"

	"gorm.io/gorm"
)

type SubscriptionService struct {
	subRepo          *repository.SubscriptionRepository
	themeContentRepo *repository.ThemeContentRepository
}

func NewSubscriptionService(subRepo *repository.SubscriptionRepository, tcRepo *repository.ThemeContentRepository) *SubscriptionService {
	return &SubscriptionService{subRepo: subRepo, themeContentRepo: tcRepo}
}

type SubscriptionResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

type ToggleEnabledRequest struct {
	Enabled *bool `json:"enabled" vd:"$!=nil"`
}

type AppendDescriptionRequest struct {
	FormerName string `json:"former_name" vd:"len($)>0"`
}

func toSubscriptionResponse(s *model.XqSubscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:          strconv.FormatInt(s.ID, 10),
		UserID:      strconv.FormatInt(s.UserID, 10),
		Description: s.Description,
		Enabled:     s.Enabled,
	}
}

func (s *SubscriptionService) Create(sub *model.XqSubscription) (*model.XqSubscription, error) {
	if sub.UserID == 0 {
		return nil, errcode.ErrBadRequest
	}
	sub.ID = snowflake.NextID()
	if err := s.subRepo.Create(sub); err != nil {
		return nil, errcode.ErrInternal
	}
	return sub, nil
}

func (s *SubscriptionService) Delete(id, userID int64) error {
	rows, err := s.subRepo.Delete(id, userID)
	if err != nil {
		return errcode.ErrInternal
	}
	if rows == 0 {
		return errcode.ErrSubscriptionNotFound
	}
	return nil
}

func (s *SubscriptionService) GetByID(id, userID int64) (*model.XqSubscription, error) {
	sub, err := s.subRepo.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrSubscriptionNotFound
		}
		return nil, errcode.ErrInternal
	}
	return sub, nil
}

func (s *SubscriptionService) GetByUserID(userID int64) ([]model.XqSubscription, error) {
	subs, err := s.subRepo.GetByUserID(userID)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	return subs, nil
}

func (s *SubscriptionService) GetAll() ([]SubscriptionResponse, error) {
	subs, err := s.subRepo.GetAll()
	if err != nil {
		return nil, errcode.ErrInternal
	}
	result := make([]SubscriptionResponse, 0, len(subs))
	for i := range subs {
		result = append(result, toSubscriptionResponse(&subs[i]))
	}
	return result, nil
}

func (s *SubscriptionService) ToggleEnabled(userID int64, req *ToggleEnabledRequest) error {
	rows, err := s.subRepo.UpdateEnabled(userID, *req.Enabled)
	if err != nil {
		return errcode.ErrInternal
	}
	if rows == 0 {
		return errcode.ErrSubscriptionNotFound
	}
	return nil
}

func (s *SubscriptionService) Exists(id, userID int64) (bool, error) {
	exists, err := s.subRepo.Exists(id, userID)
	if err != nil {
		return false, errcode.ErrInternal
	}
	return exists, nil
}

func (s *SubscriptionService) AppendDescription(id, userID int64, req *AppendDescriptionRequest) error {
	sub, err := s.subRepo.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrSubscriptionNotFound
		}
		return errcode.ErrInternal
	}

	formerName := strings.TrimSpace(req.FormerName)
	var newDesc string
	if sub.Description == "" {
		newDesc = formerName
	} else {
		newDesc = sub.Description + "|" + formerName
	}

	if err := s.subRepo.UpdateDescription(id, userID, newDesc); err != nil {
		return errcode.ErrInternal
	}
	return nil
}

func (s *SubscriptionService) Detail(id, userID int64) (*SubscriptionResponse, error) {
	sub, err := s.subRepo.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrSubscriptionNotFound
		}
		return nil, errcode.ErrInternal
	}
	resp := toSubscriptionResponse(sub)
	return &resp, nil
}

type DetailTableResponse struct {
	PageNumber int                    `json:"pageNumber"`
	PageSize   int                    `json:"pageSize"`
	TotalPage  int                    `json:"totalPage"`
	TotalCount int64                  `json:"totalCount"`
	List       []model.XqThemeContent `json:"list"`
}

func (s *SubscriptionService) DetailTable(userID int64, pageNumber, pageSize int) (*DetailTableResponse, error) {
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (pageNumber - 1) * pageSize
	list, total, err := s.themeContentRepo.GetByUserID(userID, pageSize, offset)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}

	return &DetailTableResponse{
		PageNumber: pageNumber,
		PageSize:   pageSize,
		TotalPage:  totalPage,
		TotalCount: total,
		List:       list,
	}, nil
}

func (s *SubscriptionService) GetAllEnabledMap() (map[int64]model.XqSubscription, error) {
	subs, err := s.subRepo.GetAllEnabled()
	if err != nil {
		return nil, err
	}
	m := make(map[int64]model.XqSubscription, len(subs))
	for _, sub := range subs {
		m[sub.UserID] = sub
	}
	return m, nil
}

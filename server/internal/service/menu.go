package service

import (
	"errors"
	"strconv"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/snowflake"

	"gorm.io/gorm"
)

type MenuService struct {
	menuRepo *repository.MenuRepository
}

func NewMenuService(menuRepo *repository.MenuRepository) *MenuService {
	return &MenuService{menuRepo: menuRepo}
}

type AddMenuRequest struct {
	Title    string           `json:"title"`
	Key      string           `json:"key"`
	LinkURL  string           `json:"link_url"`
	Children []AddMenuRequest `json:"children"`
}

type RoleBindingRequest struct {
	RoleID  int64    `json:"roleId" vd:"$>0"`
	MenuIDs []string `json:"menuIds" vd:"len($)>0"`
}

func (s *MenuService) List() ([]MenuResponse, error) {
	menus, err := s.menuRepo.ListActive()
	if err != nil {
		return nil, errcode.ErrInternal
	}

	seen := make(map[string]bool)
	var filtered []model.Menu
	for _, m := range menus {
		key := m.LinkURL + "|" + m.MenuCode
		if seen[key] {
			continue
		}
		seen[key] = true
		filtered = append(filtered, m)
	}
	return ToMenuResponses(filtered), nil
}

func (s *MenuService) AddAll(items []AddMenuRequest) (int, error) {
	var menuData []model.Menu
	s.processNodes(items, &menuData, 0, 1)

	if len(menuData) == 0 {
		return 0, nil
	}

	if err := s.menuRepo.BatchCreate(menuData); err != nil {
		return 0, errcode.ErrInternal
	}
	return len(menuData), nil
}

func (s *MenuService) processNodes(items []AddMenuRequest, menuData *[]model.Menu, parentID int64, level int) {
	for _, item := range items {
		_, err := s.menuRepo.GetByLinkURLAndCode(item.LinkURL, item.Key)
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}

		id := snowflake.NextID()
		if id == 0 {
			continue
		}

		*menuData = append(*menuData, model.Menu{
			ID:       id,
			Name:     item.Title,
			LinkURL:  item.LinkURL,
			MenuCode: item.Key,
			ParentID: parentID,
			Level:    level,
			IsDelete: 0,
		})

		if len(item.Children) > 0 {
			s.processNodes(item.Children, menuData, id, level+1)
		}
	}
}

func (s *MenuService) RoleBinding(req *RoleBindingRequest, currentUserID int64) error {
	exists, err := s.menuRepo.RoleExists(req.RoleID)
	if err != nil {
		return errcode.ErrInternal
	}
	if !exists {
		return errcode.ErrRoleNotFound
	}

	menuIDs := make([]int64, 0, len(req.MenuIDs))
	for _, idStr := range req.MenuIDs {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return errcode.ErrBadRequest
		}
		menuIDs = append(menuIDs, id)
	}

	if err := s.menuRepo.DeleteMenuRoleByRoleAndMenus(req.RoleID, menuIDs); err != nil {
		return errcode.ErrInternal
	}

	now := time.Now()
	for _, menuID := range menuIDs {
		record := &model.SysMenuRole{
			MenuID:     menuID,
			RoleID:     req.RoleID,
			CreateTime: now,
			UpdateTime: now,
			CreateUser: currentUserID,
			UpdateUser: currentUserID,
			IsDelete:   0,
		}
		if err := s.menuRepo.CreateMenuRole(record); err != nil {
			return errcode.ErrInternal
		}
	}
	return nil
}

func (s *MenuService) ListByRoleID(roleID int64) ([]MenuResponse, error) {
	menus, err := s.menuRepo.GetMenusByRoleID(roleID)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	return ToMenuResponses(menus), nil
}

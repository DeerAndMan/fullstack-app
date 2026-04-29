package service

import (
	"errors"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"

	"gorm.io/gorm"
)

type RoleService struct {
	roleRepo *repository.RoleRepository
}

func NewRoleService(roleRepo *repository.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}

type CreateRoleRequest struct {
	RoleName string  `json:"role_name" vd:"len($)>0 && len($)<=64"`
	RoleKey  string  `json:"role_key" vd:"len($)>0 && len($)<=100"`
	Sort     int     `json:"sort"`
	Remark   *string `json:"remark"`
}

type UpdateRoleRequest struct {
	RoleName   string  `json:"role_name"`
	RoleKey    string  `json:"role_key"`
	Sort       *int    `json:"sort"`
	Remark     *string `json:"remark"`
	RoleStatus *int8   `json:"role_status"`
}

type ListRoleRequest struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Keyword  string `query:"keyword"`
}

func (s *RoleService) Create(req *CreateRoleRequest) error {
	codeExists, err := s.roleRepo.ExistsByCode(req.RoleKey)
	if err != nil {
		return errcode.ErrInternal
	}
	if codeExists {
		return errcode.ErrRoleCodeExists
	}

	nameExists, err := s.roleRepo.ExistsByName(req.RoleName)
	if err != nil {
		return errcode.ErrInternal
	}
	if nameExists {
		return errcode.ErrRoleNameExists
	}

	now := time.Now()
	role := &model.Role{
		RoleName:   req.RoleName,
		RoleKey:    req.RoleKey,
		Sort:       req.Sort,
		Remark:     req.Remark,
		RoleStatus: 1,
		CreateBy:   "",
		CreateTime: now,
		UpdateBy:   "",
		UpdateTime: now,
		DelFlag:    0,
	}
	return s.roleRepo.Create(role)
}

func (s *RoleService) GetByID(id uint) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrRoleNotFound
		}
		return nil, errcode.ErrInternal
	}
	return role, nil
}

func (s *RoleService) Update(id uint, req *UpdateRoleRequest) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRoleNotFound
		}
		return errcode.ErrInternal
	}

	if req.RoleName != "" {
		role.RoleName = req.RoleName
	}
	if req.RoleKey != "" {
		role.RoleKey = req.RoleKey
	}
	if req.Sort != nil {
		role.Sort = *req.Sort
	}
	if req.Remark != nil {
		role.Remark = req.Remark
	}
	if req.RoleStatus != nil {
		role.RoleStatus = *req.RoleStatus
	}
	role.UpdateTime = time.Now()

	return s.roleRepo.Update(role)
}

func (s *RoleService) Delete(id uint) error {
	return s.roleRepo.Delete(id)
}

func (s *RoleService) List(req *ListRoleRequest) ([]model.Role, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	return s.roleRepo.List(req.Page, req.PageSize, req.Keyword)
}

func (s *RoleService) GetAll() ([]model.Role, error) {
	return s.roleRepo.GetAll()
}

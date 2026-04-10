package service

import (
	"errors"

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
	Name   string `json:"name" vd:"len($)>0 && len($)<=64"`
	Code   string `json:"code" vd:"len($)>0 && len($)<=64"`
	Remark string `json:"remark"`
}

type UpdateRoleRequest struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Remark string `json:"remark"`
	Status *int8  `json:"status"`
}

type ListRoleRequest struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Keyword  string `query:"keyword"`
}

func (s *RoleService) Create(req *CreateRoleRequest) error {
	exists, err := s.roleRepo.ExistsByCode(req.Code)
	if err != nil {
		return errcode.ErrInternal
	}
	if exists {
		return errcode.ErrRoleNameExists
	}

	role := &model.Role{
		Name:   req.Name,
		Code:   req.Code,
		Remark: req.Remark,
		Status: 1,
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

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Code != "" {
		role.Code = req.Code
	}
	if req.Remark != "" {
		role.Remark = req.Remark
	}
	if req.Status != nil {
		role.Status = *req.Status
	}

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

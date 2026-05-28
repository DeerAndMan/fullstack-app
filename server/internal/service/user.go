package service

import (
	"errors"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

func NewUserService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository) *UserService {
	return &UserService{userRepo: userRepo, roleRepo: roleRepo}
}

type CreateUserRequest struct {
	Name     string `json:"name" vd:"len($)>0 && len($)<=64"`
	Password string `json:"password" vd:"len($)>=6 && len($)<=32"`
	Age      int    `json:"age"`
	Email    string `json:"email"`
}

type UpdateUserRequest struct {
	Name        string `json:"name"`
	Age         *int   `json:"age"`
	Email       string `json:"email"`
	Description string `json:"description"`
}

type ListUserRequest struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Keyword  string `query:"keyword"`
}

type UserResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Age         int    `json:"age"`
	Email       string `json:"email"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Status      int8   `json:"status"`
}

func ToUserResponse(user *model.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Age:         user.Age,
		Email:       user.Email,
		Description: user.Description,
		Status:      user.Status,
	}
}

func ToUserResponses(users []model.User) []UserResponse {
	list := make([]UserResponse, 0, len(users))
	for i := range users {
		list = append(list, *ToUserResponse(&users[i]))
	}
	return list
}

func (s *UserService) Create(req *CreateUserRequest) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errcode.ErrInternal
	}

	user := &model.User{
		Name:     req.Name,
		Password: string(hashed),
		Age:      req.Age,
		Email:    req.Email,
	}
	return s.userRepo.Create(user)
}

func (s *UserService) GetByID(id uint) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrInternal
	}
	return ToUserResponse(user), nil
}

func (s *UserService) Update(id uint, req *UpdateUserRequest) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrUserNotFound
		}
		return errcode.ErrInternal
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Age != nil {
		user.Age = *req.Age
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Description != "" {
		user.Description = req.Description
	}

	if err := s.userRepo.Update(user); err != nil {
		return errcode.ErrInternal
	}
	return nil
}

func (s *UserService) Delete(id uint) error {
	return s.userRepo.Delete(id)
}

func (s *UserService) List(req *ListUserRequest) ([]UserResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	users, total, err := s.userRepo.List(req.Page, req.PageSize, req.Keyword)
	if err != nil {
		return nil, 0, err
	}
	return ToUserResponses(users), total, nil
}

type UpdateUserRoleRequest struct {
	RoleID int64 `json:"role_id" vd:"$>0"`
}

type AssignUserRolesRequest struct {
	RoleIDs []int64 `json:"role_ids" vd:"len($)>0"`
}

type UserRoleResponse struct {
	RoleID   int64  `json:"role_id"`
	RoleName string `json:"role_name"`
	RoleKey  string `json:"role_key"`
}

func (s *UserService) UpdateRole(userID uint, req *UpdateUserRoleRequest) (*model.Role, error) {
	if _, err := s.userRepo.GetByID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrInternal
	}

	role, err := s.roleRepo.GetByID(uint(req.RoleID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrRoleNotFound
		}
		return nil, errcode.ErrInternal
	}
	if role.RoleStatus != 0 {
		return nil, errcode.ErrRoleDisabled
	}

	if err := s.userRepo.DeleteUserRoles(userID); err != nil {
		return nil, errcode.ErrInternal
	}
	ur := &model.SysUserRole{UserID: int64(userID), RoleID: req.RoleID}
	if err := s.userRepo.CreateUserRole(ur); err != nil {
		return nil, errcode.ErrInternal
	}
	return role, nil
}

func (s *UserService) AssignRoles(userID uint, req *AssignUserRolesRequest) ([]model.Role, error) {
	if _, err := s.userRepo.GetByID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrInternal
	}

	var roles []model.Role
	for _, rid := range req.RoleIDs {
		role, err := s.roleRepo.GetByID(uint(rid))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrRoleInvalid
			}
			return nil, errcode.ErrInternal
		}
		if role.RoleStatus != 0 {
			return nil, errcode.ErrRoleDisabled
		}
		roles = append(roles, *role)
	}

	if err := s.userRepo.DeleteUserRoles(userID); err != nil {
		return nil, errcode.ErrInternal
	}
	urs := make([]model.SysUserRole, 0, len(req.RoleIDs))
	for _, rid := range req.RoleIDs {
		urs = append(urs, model.SysUserRole{UserID: int64(userID), RoleID: rid})
	}
	if err := s.userRepo.BatchCreateUserRoles(urs); err != nil {
		return nil, errcode.ErrInternal
	}
	return roles, nil
}

func (s *UserService) GetRoles(userID uint) ([]model.Role, error) {
	if _, err := s.userRepo.GetByID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrInternal
	}
	roles, err := s.userRepo.GetUserRoles(userID)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	return roles, nil
}

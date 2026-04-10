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
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type CreateUserRequest struct {
	Username string `json:"username" vd:"len($)>0 && len($)<=64"`
	Password string `json:"password" vd:"len($)>=6 && len($)<=32"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RoleIDs  []uint `json:"role_ids"`
}

type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Status   *int8  `json:"status"`
	RoleIDs  []uint `json:"role_ids"`
}

type ListUserRequest struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Keyword  string `query:"keyword"`
}

func (s *UserService) Create(req *CreateUserRequest) error {
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return errcode.ErrInternal
	}
	if exists {
		return errcode.ErrUsernameExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errcode.ErrInternal
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashed),
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1,
	}
	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	if len(req.RoleIDs) > 0 {
		return s.userRepo.AssignRoles(user.ID, req.RoleIDs)
	}
	return nil
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrInternal
	}
	return user, nil
}

func (s *UserService) Update(id uint, req *UpdateUserRequest) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrUserNotFound
		}
		return errcode.ErrInternal
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.userRepo.Update(user); err != nil {
		return errcode.ErrInternal
	}

	if req.RoleIDs != nil {
		return s.userRepo.AssignRoles(id, req.RoleIDs)
	}
	return nil
}

func (s *UserService) Delete(id uint) error {
	return s.userRepo.Delete(id)
}

func (s *UserService) List(req *ListUserRequest) ([]model.User, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	return s.userRepo.List(req.Page, req.PageSize, req.Keyword)
}

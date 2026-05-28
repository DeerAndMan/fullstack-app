package service

import (
	"errors"
	"strconv"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"
	jwtpkg "fullstack-app/server/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *jwtpkg.Manager
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *jwtpkg.Manager) *AuthService {
	return &AuthService{userRepo: userRepo, jwtManager: jwtManager}
}

type RegisterRequest struct {
	Name     string `json:"name" vd:"len($)>0 && len($)<=64"`
	Password string `json:"password" vd:"len($)>0"`
}

type LoginRequest struct {
	Name     string `json:"name" vd:"len($)>0"`
	Password string `json:"password" vd:"len($)>0"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" vd:"len($)>0"`
}

func (s *AuthService) Register(req *RegisterRequest) error {
	exists, err := s.userRepo.ExistsByName(req.Name)
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
		Name:     req.Name,
		Password: string(hashed),
	}
	return s.userRepo.Create(user)
}

type LoginResponse struct {
	Token     *jwtpkg.TokenPair `json:"token"`
	User      *UserResponse     `json:"user"`
	Role      *model.Role       `json:"role"`
	MenuRoles []MenuResponse    `json:"menuRoles"`
}

type MenuResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LinkURL  string `json:"link_url"`
	MenuCode string `json:"menu_code"`
	ParentID string `json:"parent_id"`
	NodeType int8   `json:"node_type"`
	IconURL  string `json:"icon_url"`
	Level    int    `json:"level"`
	Path     string `json:"path"`
	IsDelete int8   `json:"is_delete"`
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.GetByName(req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInvalidCredentials
		}
		return nil, errcode.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errcode.ErrInvalidCredentials
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Name)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	userResp := ToUserResponse(user)
	if userResp != nil {
		avatar, err := s.userRepo.GetLatestAvatar(user.ID)
		if err != nil {
			return nil, errcode.ErrInternal
		}
		userResp.Avatar = avatar
	}

	role, err := s.userRepo.GetRoleByUserID(user.ID)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	var menuRoles []MenuResponse
	if role != nil {
		menus, err := s.userRepo.GetMenusByRoleID(role.RoleID)
		if err != nil {
			return nil, errcode.ErrInternal
		}
		menuRoles = ToMenuResponses(menus)
	}

	return &LoginResponse{
		Token:     tokenPair,
		User:      userResp,
		Role:      role,
		MenuRoles: menuRoles,
	}, nil
}

func (s *AuthService) RefreshToken(req *RefreshTokenRequest) (*jwtpkg.TokenPair, error) {
	claims, err := s.jwtManager.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errcode.ErrTokenInvalid
	}

	return s.jwtManager.GenerateTokenPair(claims.UserID, claims.Username)
}

func (s *AuthService) Logout() error {
	// 当前 JWT 为无状态令牌，服务端不维护会话
	// 后续可接入 Redis 黑名单机制
	return nil
}

func ToMenuResponses(menus []model.Menu) []MenuResponse {
	list := make([]MenuResponse, 0, len(menus))
	for _, menu := range menus {
		list = append(list, MenuResponse{
			ID:       strconv.FormatInt(menu.ID, 10),
			Name:     menu.Name,
			LinkURL:  menu.LinkURL,
			MenuCode: menu.MenuCode,
			ParentID: strconv.FormatInt(menu.ParentID, 10),
			NodeType: menu.NodeType,
			IconURL:  menu.IconURL,
			Level:    menu.Level,
			Path:     menu.Path,
			IsDelete: menu.IsDelete,
		})
	}
	return list
}

package service

import (
	"errors"
	"strconv"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"
	jwtpkg "fullstack-app/server/pkg/jwt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *jwtpkg.Manager
	rdb        *redis.Client // 预留：当前会话读写已下沉到 jwtManager，业务层暂未直接使用
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *jwtpkg.Manager, rdb *redis.Client) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		rdb:        rdb,
	}
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

	// 生成不透明令牌对（会话信息由 jwtManager 写入 Redis）
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, time.Now().Unix())
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

// RefreshToken 校验 refresh 令牌并签发新的令牌对。
// 注意：旧令牌对应的 Redis 会话不会主动清除，会随各自 TTL 自然过期
func (s *AuthService) RefreshToken(req *RefreshTokenRequest) (*jwtpkg.TokenPair, error) {
	claims, err := s.jwtManager.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errcode.ErrTokenInvalid
	}

	// 刷新时保留原会话的登录时间，语义上仍是同一次登录
	return s.jwtManager.GenerateTokenPair(claims.UserID, claims.LoginAt)
}

func (s *AuthService) Logout(accessToken string) error {
	// 不透明令牌：登出即删除 Redis 中的 access 会话，实现主动吊销
	if err := s.jwtManager.Revoke(accessToken); err != nil {
		zap.L().Error("吊销令牌失败", zap.Error(err))
		return errcode.ErrInternal
	}
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

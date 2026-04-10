package service

import (
	"errors"

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
	Username string `json:"username" vd:"len($)>0 && len($)<=64"`
	Password string `json:"password" vd:"len($)>=6 && len($)<=32"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Username string `json:"username" vd:"len($)>0"`
	Password string `json:"password" vd:"len($)>0"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" vd:"len($)>0"`
}

func (s *AuthService) Register(req *RegisterRequest) error {
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
		Status:   1,
	}
	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	return s.userRepo.Create(user)
}

func (s *AuthService) Login(req *LoginRequest) (*jwtpkg.TokenPair, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInvalidCredentials
		}
		return nil, errcode.ErrInternal
	}

	if user.Status == 0 {
		return nil, errcode.ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errcode.ErrInvalidCredentials
	}

	return s.jwtManager.GenerateTokenPair(user.ID, user.Username)
}

func (s *AuthService) RefreshToken(req *RefreshTokenRequest) (*jwtpkg.TokenPair, error) {
	claims, err := s.jwtManager.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, errcode.ErrTokenInvalid
	}

	return s.jwtManager.GenerateTokenPair(claims.UserID, claims.Username)
}

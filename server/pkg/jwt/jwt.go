package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type Manager struct {
	secret        []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
	issuer        string
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func NewManager(secret string, accessExpireHours, refreshExpireHours float64, issuer string) *Manager {
	return &Manager{
		secret:        []byte(secret),
		accessExpire:  time.Duration(accessExpireHours * float64(time.Hour)),
		refreshExpire: time.Duration(refreshExpireHours * float64(time.Hour)),
		issuer:        issuer,
	}
}

func (m *Manager) GenerateTokenPair(userID uint, username string) (*TokenPair, error) {
	now := time.Now()
	accessExpAt := now.Add(m.accessExpire)

	accessToken, err := m.generateToken(userID, username, "access", accessExpAt)
	if err != nil {
		return nil, err
	}

	refreshToken, err := m.generateToken(userID, username, "refresh", now.Add(m.refreshExpire))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpAt.Unix(),
	}, nil
}

func (m *Manager) generateToken(userID uint, username string, tokenType string, expiresAt time.Time) (string, error) {
	claims := Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	claims, err := m.ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type: expected access token")
	}
	return claims, nil
}

func (m *Manager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	claims, err := m.ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type: expected refresh token")
	}
	return claims, nil
}

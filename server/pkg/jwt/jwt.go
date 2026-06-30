package jwt

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Claims 解析令牌后得到的会话声明（从 Redis 反序列化而来）
type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"` // "access" or "refresh"
}

// sessionData Redis 中实际存储的会话详情。
// 令牌本身是不透明随机串，所有用户信息都保存在这里。
type sessionData struct {
	UserID    uint   `json:"uid"`
	Username  string `json:"name"`
	ExpiresAt int64  `json:"exp"` // 逻辑过期时间（Unix 秒），用于区分 expired 与 invalid
}

type Manager struct {
	rdb           *redis.Client
	accessExpire  time.Duration
	refreshExpire time.Duration
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

const (
	accessKeyPrefix  = "session:a:" // access 令牌会话前缀
	refreshKeyPrefix = "session:r:" // refresh 令牌会话前缀
)

// NewManager 构造令牌管理器。令牌信息存储于 Redis，令牌本身仅为不透明随机串。
func NewManager(rdb *redis.Client, accessExpireHours, refreshExpireHours float64) *Manager {
	return &Manager{
		rdb:           rdb,
		accessExpire:  time.Duration(accessExpireHours * float64(time.Hour)),
		refreshExpire: time.Duration(refreshExpireHours * float64(time.Hour)),
	}
}

// randToken 生成 32 字节随机串并 base64url 编码（约 43 字符，不透明、不含用户信息）。
func randToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateTokenPair 为给定用户生成新的访问和刷新令牌对，并把会话写入 Redis。
func (m *Manager) GenerateTokenPair(userID uint, username string) (*TokenPair, error) {
	ctx := context.Background()
	now := time.Now()
	accessExpAt := now.Add(m.accessExpire)
	refreshExpAt := now.Add(m.refreshExpire)

	accessToken, err := randToken()
	if err != nil {
		return nil, err
	}
	refreshToken, err := randToken()
	if err != nil {
		return nil, err
	}

	// access 会话：Redis key 存活到 refresh 过期，便于区分 expired/invalid；
	// value 中记录逻辑过期时间 exp，解析时据此判断是否过期。
	accessVal, err := json.Marshal(sessionData{UserID: userID, Username: username, ExpiresAt: accessExpAt.Unix()})
	if err != nil {
		return nil, err
	}
	if err := m.rdb.Set(ctx, accessKeyPrefix+accessToken, accessVal, m.refreshExpire).Err(); err != nil {
		return nil, fmt.Errorf("write access session: %w", err)
	}

	// refresh 会话：TTL 即为 refresh 过期时长。
	refreshVal, err := json.Marshal(sessionData{UserID: userID, Username: username, ExpiresAt: refreshExpAt.Unix()})
	if err != nil {
		return nil, err
	}
	if err := m.rdb.Set(ctx, refreshKeyPrefix+refreshToken, refreshVal, m.refreshExpire).Err(); err != nil {
		return nil, fmt.Errorf("write refresh session: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpAt.Unix(),
	}, nil
}

// parse 从 Redis 读取会话并校验有效性。
// key 不存在 -> invalid；存在但逻辑过期时间已过 -> expired。
func (m *Manager) parse(keyPrefix, tokenStr, tokenType string) (*Claims, error) {
	val, err := m.rdb.Get(context.Background(), keyPrefix+tokenStr).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("invalid token")
		}
		return nil, fmt.Errorf("read session: %w", err)
	}

	var sd sessionData
	if err := json.Unmarshal(val, &sd); err != nil {
		return nil, errors.New("invalid token")
	}
	if time.Now().Unix() > sd.ExpiresAt {
		return nil, errors.New("token expired")
	}

	return &Claims{UserID: sd.UserID, Username: sd.Username, TokenType: tokenType}, nil
}

func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	return m.parse(accessKeyPrefix, tokenStr, "access")
}

func (m *Manager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return m.parse(refreshKeyPrefix, tokenStr, "refresh")
}

// Revoke 主动吊销 access 令牌（用于登出），删除其 Redis 会话。
func (m *Manager) Revoke(accessToken string) error {
	if accessToken == "" {
		return nil
	}
	return m.rdb.Del(context.Background(), accessKeyPrefix+accessToken).Err()
}

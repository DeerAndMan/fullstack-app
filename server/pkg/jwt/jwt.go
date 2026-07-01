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
	LoginAt   int64  `json:"login_at"`   // 登录时间（Unix 秒）
	ExpiresAt int64  `json:"expires_at"` // 逻辑过期时间（Unix 秒），用于判断是否临近过期
	TokenType string `json:"token_type"` // "access" or "refresh"
}

// sessionData Redis 中实际存储的会话详情。
// 令牌本身是不透明随机串，所有用户信息都保存在这里。
type sessionData struct {
	UserID    uint  `json:"uid"`
	LoginAt   int64 `json:"login_at"`   // 登录时间（Unix 秒）
	ExpiresAt int64 `json:"expires_at"` // 逻辑过期时间（Unix 秒），用于区分 expired 与 invalid
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
	accessKeyPrefix  = "session:access:"  // access 令牌会话前缀
	refreshKeyPrefix = "session:refresh:" // refresh 令牌会话前缀

	// renewThreshold 滑动续期阈值：access 剩余有效期不足总时长该比例时签发新令牌。
	renewThreshold = 0.3
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
// loginAt 为登录时间（Unix 秒）：首次登录传当前时间，刷新令牌时传原会话的登录时间以保留语义。
func (m *Manager) GenerateTokenPair(userID uint, loginAt int64) (*TokenPair, error) {
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
	accessVal, err := json.Marshal(sessionData{UserID: userID, LoginAt: loginAt, ExpiresAt: accessExpAt.Unix()})
	if err != nil {
		return nil, err
	}
	if err := m.rdb.Set(ctx, accessKeyPrefix+accessToken, accessVal, m.refreshExpire).Err(); err != nil {
		return nil, fmt.Errorf("write access session: %w", err)
	}

	// refresh 会话：TTL 即为 refresh 过期时长。
	refreshVal, err := json.Marshal(sessionData{UserID: userID, LoginAt: loginAt, ExpiresAt: refreshExpAt.Unix()})
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

	return &Claims{UserID: sd.UserID, LoginAt: sd.LoginAt, ExpiresAt: sd.ExpiresAt, TokenType: tokenType}, nil
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

// issueAccessToken 仅签发一个新的 access 令牌并写入 Redis（不动 refresh）。
// 用于滑动续期：续期只延长 access 生命周期，refresh 保持不变。
func (m *Manager) issueAccessToken(userID uint, loginAt int64) (string, int64, error) {
	now := time.Now()
	accessExpAt := now.Add(m.accessExpire)

	token, err := randToken()
	if err != nil {
		return "", 0, err
	}
	val, err := json.Marshal(sessionData{UserID: userID, LoginAt: loginAt, ExpiresAt: accessExpAt.Unix()})
	if err != nil {
		return "", 0, err
	}
	// 与 GenerateTokenPair 保持一致：access 会话 key 存活到 refresh 过期，
	// 逻辑过期时间记录在 value 里。
	if err := m.rdb.Set(context.Background(), accessKeyPrefix+token, val, m.refreshExpire).Err(); err != nil {
		return "", 0, fmt.Errorf("write access session: %w", err)
	}
	return token, accessExpAt.Unix(), nil
}

// MaybeRenewAccessToken 在 access 令牌临近过期时签发一个新令牌用于无感续期。
// 当剩余有效期不足 access 总时长的 renewThreshold 比例时才续期。
// 返回空串表示无需续期。旧令牌不主动吊销，交由其自然过期，避免前端并发请求竞态。
func (m *Manager) MaybeRenewAccessToken(claims *Claims) (newToken string, err error) {
	if claims == nil || claims.TokenType != "access" {
		return "", nil
	}
	// 剩余有效期（秒）
	remaining := claims.ExpiresAt - time.Now().Unix()
	if remaining <= 0 {
		return "", nil // 已过期不会走到这里，兜底
	}
	// 阈值：剩余不足总时长的 30% 时续期
	threshold := int64(m.accessExpire.Seconds() * renewThreshold)
	if remaining > threshold {
		return "", nil
	}
	token, _, err := m.issueAccessToken(claims.UserID, claims.LoginAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

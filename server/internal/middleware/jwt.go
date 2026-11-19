package middleware

import (
	"context"
	"net/http"
	"strings"

	"fullstack-app/server/pkg/errcode"
	jwtpkg "fullstack-app/server/pkg/jwt"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

const (
	HeaderAuthorization = "Authorization"
	HeaderNewToken      = "x-new-token" // 滑动续期：临近过期时通过该响应头下发新 access 令牌
	BearerPrefix        = "Bearer "
	CtxUserIDKey        = "user_id"
	CtxLoginAtKey       = "login_at"
)

// JWTAuth 鉴权中间件：校验 Authorization 头中的不透明令牌，
// 通过 jwtManager 查 Redis 还原会话，并把用户信息写入请求上下文
func JWTAuth(jwtManager *jwtpkg.Manager) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		auth := string(c.GetHeader(HeaderAuthorization))
		if auth == "" || !strings.HasPrefix(auth, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Body{
				Code:    errcode.ErrUnauthorized.Code,
				Message: errcode.ErrUnauthorized.Message,
			})
			return
		}

		tokenStr := strings.TrimPrefix(auth, BearerPrefix)
		claims, err := jwtManager.ParseAccessToken(tokenStr)
		// fmt.Printf("当前用户返回的TOKEN详细数据： %+v \n", *claims)

		if err != nil {
			code := errcode.ErrTokenInvalid
			if strings.Contains(err.Error(), "expired") {
				code = errcode.ErrTokenExpired
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Body{
				Code:    code.Code,
				Message: code.Message,
			})
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxLoginAtKey, claims.LoginAt)

		// 滑动续期：access 令牌临近过期时签发新令牌，通过响应头下发。
		// 续期失败不阻断请求，用户仍可用当前令牌继续访问直至其自然过期。
		if newToken, rErr := jwtManager.MaybeRenewAccessToken(claims); rErr == nil && newToken != "" {
			// fmt.Printf("新的token的值： %+v \n", newToken)
			c.Header(HeaderNewToken, newToken)
		}

		c.Next(ctx)
	}
}

func GetUserID(c *app.RequestContext) uint {
	v, _ := c.Get(CtxUserIDKey)
	id, _ := v.(uint)
	return id
}

func GetLoginAt(c *app.RequestContext) int64 {
	v, _ := c.Get(CtxLoginAtKey)
	t, _ := v.(int64)
	return t
}

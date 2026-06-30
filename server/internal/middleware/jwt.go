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
	BearerPrefix        = "Bearer "
	CtxUserIDKey        = "user_id"
	CtxUsernameKey      = "username"
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
		c.Set(CtxUsernameKey, claims.Username)
		c.Next(ctx)
	}
}

func GetUserID(c *app.RequestContext) uint {
	v, _ := c.Get(CtxUserIDKey)
	id, _ := v.(uint)
	return id
}

func GetUsername(c *app.RequestContext) string {
	v, _ := c.Get(CtxUsernameKey)
	s, _ := v.(string)
	return s
}

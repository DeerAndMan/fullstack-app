package middleware

import (
	"context"
	"net/http"

	"fullstack-app/server/pkg/response"

	"github.com/casbin/casbin/v2"
	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

func CasbinRBAC(enforcer *casbin.Enforcer) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		username := GetUsername(c)
		path := string(c.Request.URI().Path())
		method := string(c.Method())

		ok, err := enforcer.Enforce(username, path, method)
		if err != nil {
			zap.L().Error("casbin enforce error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Body{
				Code:    500,
				Message: "internal server error",
			})
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Body{
				Code:    403,
				Message: "forbidden",
			})
			return
		}

		c.Next(ctx)
	}
}

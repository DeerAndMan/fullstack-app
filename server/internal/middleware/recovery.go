package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime"

	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				zap.L().Error("panic recovered",
					zap.String("error", fmt.Sprintf("%v", r)),
					zap.String("stack", string(buf[:n])),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Body{
					Code:    errcode.ErrInternal.Code,
					Data:    nil,
					Message: errcode.ErrInternal.Message,
				})
			}
		}()
		c.Next(ctx)
	}
}

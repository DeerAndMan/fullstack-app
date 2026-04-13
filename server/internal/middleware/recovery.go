package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime"

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
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"data":    nil,
					"message": "internal server error",
				})
			}
		}()
		c.Next(ctx)
	}
}

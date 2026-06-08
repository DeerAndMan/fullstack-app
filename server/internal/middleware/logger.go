package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

func Logger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		method := string(c.Method())

		c.Next(ctx)

		latency := time.Since(start)
		status := c.Response.StatusCode()

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		}

		if v, exists := c.Get("request_id"); exists {
			if reqID, ok := v.(string); ok && reqID != "" {
				fields = append(fields, zap.String("request_id", reqID))
			}
		}

		if status >= 500 {
			zap.L().Error("request", fields...)
		} else if status >= 400 {
			zap.L().Warn("request", fields...)
		} else {
			zap.L().Info("request", fields...)
		}
	}
}

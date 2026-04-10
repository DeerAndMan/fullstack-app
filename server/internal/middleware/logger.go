package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

func Logger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		method := string(c.Method())

		c.Next(ctx)

		latency := time.Since(start)
		status := c.Response.StatusCode()

		attrs := []any{
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
		}

		if reqID := string(c.GetHeader("X-Request-ID")); reqID != "" {
			attrs = append(attrs, slog.String("request_id", reqID))
		}

		if status >= 500 {
			slog.Error("request", attrs...)
		} else if status >= 400 {
			slog.Warn("request", attrs...)
		} else {
			slog.Info("request", attrs...)
		}
	}
}

package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

func CORS() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		allowOrigins := map[string]bool{
			"http://localhost:6565":  true,
			"http://127.0.0.1:6565": true,
			"http://localhost:7878":  true,
			"http://127.0.0.1:7878": true,
		}
		if allowOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, x-new-token")
		c.Header("Access-Control-Max-Age", "86400")

		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next(ctx)
	}
}

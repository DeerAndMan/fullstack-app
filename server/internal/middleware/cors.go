package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

func CORS(allowOrigins []string) app.HandlerFunc {
	originSet := make(map[string]bool, len(allowOrigins))
	for _, o := range allowOrigins {
		originSet[o] = true
	}

	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		if originSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID, saltlength")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, x-new-token")
		c.Header("Access-Control-Max-Age", "86400")

		// Chrome 私有网络访问（PNA）：公网页面请求 localhost 时，预检会带
		// Access-Control-Request-Private-Network: true，必须回应此头才放行
		if string(c.GetHeader("Access-Control-Request-Private-Network")) == "true" {
			c.Header("Access-Control-Allow-Private-Network", "true")
		}

		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next(ctx)
	}
}

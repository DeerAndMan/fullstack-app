package router

import (
	"context"

	v1 "fullstack-app/server/internal/router/v1"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// RegisterPublicRoutes 注册无需认证的独立路由（不走 /api/v1 前缀）
func RegisterPublicRoutes(h *server.Hertz, v1Handlers *v1.Handlers) {
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	// 油猴脚本直连
	h.POST("/energy/asset", func(ctx context.Context, c *app.RequestContext) {
		v1Handlers.Energy.InsertAssets(ctx, c)
	})
}

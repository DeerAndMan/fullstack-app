package router

import (
	"context"

	"fullstack-app/server/internal/middleware"
	v1 "fullstack-app/server/internal/router/v1"
	jwtpkg "fullstack-app/server/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func Setup(h *server.Hertz, v1Handlers *v1.Handlers, jwtManager *jwtpkg.Manager, allowOrigins []string) {
	h.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.CORS(allowOrigins),
		middleware.Logger(),
	)

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	// --- v1 ---
	apiV1 := h.Group("/api/v1")    // 公开组，无 JWT
	protectedV1 := apiV1.Group("") // 受保护组，带 JWT
	protectedV1.Use(middleware.JWTAuth(jwtManager))
	v1.RegisterRoutes(apiV1, protectedV1, v1Handlers)

	// --- v2 ---
	// apiV2 := h.Group("/api/v2")
	// protectedV2 := apiV2.Group("")
	// protectedV2.Use(middleware.JWTAuth(jwtManager))
	// v2.RegisterRoutes(apiV2, protectedV2, v2Handlers)
}

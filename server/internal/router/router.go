package router

import (
	"context"

	"fullstack-app/server/internal/handler"
	"fullstack-app/server/internal/middleware"
	jwtpkg "fullstack-app/server/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

type Handlers struct {
	Auth   *handler.AuthHandler
	User   *handler.UserHandler
	Role   *handler.RoleHandler
	Upload *handler.UploadHandler
}

func Setup(h *server.Hertz, handlers *Handlers, jwtManager *jwtpkg.Manager, allowOrigins []string) {
	h.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.CORS(allowOrigins),
		middleware.Logger(),
	)

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	api := h.Group("/api/v1")

	protected := api.Group("")
	// protected.Use(middleware.JWTAuth(jwtManager))

	registerAuthRoutes(api, protected, handlers.Auth)
	registerUserRoutes(protected, handlers.User)
	registerRoleRoutes(protected, handlers.Role)
	registerUploadRoutes(protected, handlers.Upload)
}

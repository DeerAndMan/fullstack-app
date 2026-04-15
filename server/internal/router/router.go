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

func Setup(h *server.Hertz, handlers *Handlers, jwtManager *jwtpkg.Manager) {
	// Global middleware
	h.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.CORS(),
		middleware.Logger(),
	)

	// Health check
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	api := h.Group("/api/v1")

	// Public routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Auth.Register)
		auth.POST("/login", handlers.Auth.Login)
		auth.POST("/refresh-token", handlers.Auth.RefreshToken)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(jwtManager))
	{
		// Auth (protected)
		protected.POST("/auth/logout", handlers.Auth.Logout)
		// User
		users := protected.Group("/users")
		{
			users.GET("", handlers.User.List)
			users.POST("", handlers.User.Create)
			users.GET("/profile", handlers.User.GetProfile)
			users.GET("/:id", handlers.User.GetByID)
			users.PUT("/:id", handlers.User.Update)
			users.DELETE("/:id", handlers.User.Delete)
		}

		// Role
		roles := protected.Group("/roles")
		{
			roles.GET("", handlers.Role.List)
			roles.GET("/all", handlers.Role.GetAll)
			roles.POST("", handlers.Role.Create)
			roles.GET("/:id", handlers.Role.GetByID)
			roles.PUT("/:id", handlers.Role.Update)
			roles.DELETE("/:id", handlers.Role.Delete)
		}

		// Upload
		protected.POST("/upload", handlers.Upload.Upload)
	}
}

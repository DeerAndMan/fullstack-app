package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerUserRoutes(protected *route.RouterGroup, h *handler.UserHandler) {
	users := protected.Group("/users")
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.GET("/profile", h.GetProfile)
		users.GET("/:id", h.GetByID)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
		users.PUT("/:id/role", h.UpdateRole)
		users.POST("/:id/roles", h.AssignRoles)
		users.GET("/:id/roles", h.GetRoles)
	}
}

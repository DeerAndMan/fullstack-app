package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerUserRoutes(protected *route.RouterGroup, h *handlerv1.UserHandler) {
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

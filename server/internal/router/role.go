package router

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerRoleRoutes(protected *route.RouterGroup, h *handler.RoleHandler) {
	roles := protected.Group("/roles")
	{
		roles.GET("", h.List)
		roles.GET("/all", h.GetAll)
		roles.POST("", h.Create)
		roles.GET("/:id", h.GetByID)
		roles.PUT("/:id", h.Update)
		roles.DELETE("/:id", h.Delete)
	}
}

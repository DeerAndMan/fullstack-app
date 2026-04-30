package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerMenuRoutes(protected *route.RouterGroup, h *handler.MenuHandler) {
	menus := protected.Group("/menus")
	{
		menus.GET("", h.List)
		menus.POST("", h.AddAll)
		menus.POST("/role-binding", h.RoleBinding)
		menus.GET("/role-binding/:roleId", h.ListByRoleID)
	}
}

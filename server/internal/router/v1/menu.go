package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerMenuRoutes(protected *route.RouterGroup, h *handlerv1.MenuHandler) {
	menus := protected.Group("/menus")
	{
		menus.GET("", h.List)
		menus.POST("", h.AddAll)
		menus.POST("/role-binding", h.RoleBinding)
		menus.GET("/role-binding/:roleId", h.ListByRoleID)
	}
}

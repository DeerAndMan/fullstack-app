package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerRoleRoutes(protected *route.RouterGroup, h *handlerv1.RoleHandler) {
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

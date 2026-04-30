package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerEnumRoutes(public *route.RouterGroup, h *handler.EnumHandler) {
	enums := public.Group("/enums")
	{
		enums.GET("/roles", h.Roles)
	}
}

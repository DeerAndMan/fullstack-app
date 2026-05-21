package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerEnumRoutes(public *route.RouterGroup, h *handlerv1.EnumHandler) {
	enums := public.Group("/enums")
	{
		enums.GET("/roles", h.Roles)
	}
}

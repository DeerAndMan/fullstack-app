package v2

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerTestRoutes(public *route.RouterGroup, h *handler.TestHandlerV2) {
	public.GET("/test/ping", h.Ping)
}

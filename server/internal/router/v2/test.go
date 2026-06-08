package v2

import (
	handlerv2 "fullstack-app/server/internal/handler/v2"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerTestRoutes(public *route.RouterGroup, h *handlerv2.TestHandler) {
	public.GET("/test/ping", h.Ping)
}

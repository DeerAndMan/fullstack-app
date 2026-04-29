package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerSseRoutes(public *route.RouterGroup, h *handler.SseHandler) {
	sse := public.Group("/sse")
	{
		sse.POST("/chat-messages", h.ChatMessages)
		sse.GET("/chat-messages/:id", h.GetChatMessages)
	}
}

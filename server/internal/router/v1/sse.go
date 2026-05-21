package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerSseRoutes(public *route.RouterGroup, h *handlerv1.SseHandler) {
	sse := public.Group("/sse")
	{
		sse.POST("/chat-messages", h.ChatMessages)
		sse.GET("/chat-messages/:id", h.GetChatMessages)
	}
}

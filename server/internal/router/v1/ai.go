package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerAiRoutes(public *route.RouterGroup, h *handler.AiHandler) {
	ai := public.Group("/ai")
	{
		ai.GET("/conversations", h.GetConversations)
	}
}

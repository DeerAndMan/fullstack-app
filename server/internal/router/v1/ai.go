package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerAiRoutes(public *route.RouterGroup, h *handlerv1.AiHandler) {
	ai := public.Group("/ai")
	{
		ai.GET("/conversations", h.GetConversations)
	}
}

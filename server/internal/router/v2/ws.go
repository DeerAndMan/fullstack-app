package v2

import (
	handlerv2 "fullstack-app/server/internal/handler/v2"

	"github.com/cloudwego/hertz/pkg/route"
)

// registerWsRoutes 注册 V2 WebSocket 接入点。
// 前端通过 ws(s)://<host>/api/v2/ws/conversations 建立长连接。
func registerWsRoutes(public *route.RouterGroup, h *handlerv2.WsHandler) {
	ws := public.Group("/ws")
	{
		ws.GET("/conversations", h.Conversations)
	}
}

package v2

import (
	"context"

	"fullstack-app/server/internal/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/hertz-contrib/websocket"
	"go.uber.org/zap"
)

// WsHandler 提供 V2 版本的 WebSocket 接入点。
// 配合 service.WsHub 实现服务端主动推送。
type WsHandler struct {
	hub      *service.WsHub
	upgrader *websocket.HertzUpgrader
}

func NewWsHandler(hub *service.WsHub) *WsHandler {
	return &WsHandler{
		hub: hub,
		upgrader: &websocket.HertzUpgrader{
			// 开发期允许所有来源，生产请按 cfg.CORS.AllowOrigins 校验
			CheckOrigin: func(_ *app.RequestContext) bool { return true },
		},
	}
}

// Conversations 对应 GET /api/v2/ws/conversations
// 客户端建立 WS 后由 Hub 接管推送，本端只负责读循环以感知断开。
func (h *WsHandler) Conversations(_ context.Context, c *app.RequestContext) {
	connID := uuid.NewString()

	err := h.upgrader.Upgrade(c, func(conn *websocket.Conn) {
		h.hub.Register(connID, conn)
		defer h.hub.Unregister(connID)

		zap.L().Info("ws connected", zap.String("conn_id", connID), zap.Int("online", h.hub.Count()))

		// 读循环：客户端断开或网络错误时退出，触发 defer 中的注销
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				zap.L().Info("ws disconnected", zap.String("conn_id", connID), zap.Error(err))
				return
			}
			// 当前场景为纯主动推送，收到客户端消息暂不处理
		}
	})
	if err != nil {
		zap.L().Warn("ws upgrade failed", zap.Error(err))
		return
	}
}

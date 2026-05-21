package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerTradeRoutes(protected *route.RouterGroup, h *handlerv1.TradeHandler) {
	trade := protected.Group("/trade")
	{
		trade.POST("/index", h.Index)
		trade.POST("/summary", h.Summary)
	}
}

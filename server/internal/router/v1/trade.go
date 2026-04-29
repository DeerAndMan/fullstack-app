package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerTradeRoutes(protected *route.RouterGroup, h *handler.TradeHandler) {
	trade := protected.Group("/trade")
	{
		trade.POST("/index", h.Index)
		trade.POST("/summary", h.Summary)
	}
}

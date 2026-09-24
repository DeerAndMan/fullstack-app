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

// registerTradeStatsRoutes 注册交易汇总统计路由。
func registerTradeStatsRoutes(protected *route.RouterGroup, h *handlerv1.TradeStatsHandler) {
	stats := protected.Group("/trade/stats")
	{
		stats.POST("/query", h.Query) // 按 日/周/月/半年/年/自定义区间 查询汇总
		stats.POST("/sync", h.Sync)   // 从 summary 回填日终快照
		stats.GET("/range", h.Range)  // 日终快照的可用日期范围
	}
}

package v1

import (
	"context"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type TradeStatsHandler struct {
	statsSvc *service.TradeStatsService
}

func NewTradeStatsHandler(statsSvc *service.TradeStatsService) *TradeStatsHandler {
	return &TradeStatsHandler{statsSvc: statsSvc}
}

// Query 按粒度 + 自定义区间查询汇总统计。
func (h *TradeStatsHandler) Query(ctx context.Context, c *app.RequestContext) {
	var req service.StatsQueryRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	data, err := h.statsSvc.QueryStats(&req)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, data)
}

// Range 返回日终快照的可用日期范围，供前端设置默认查询区间。
func (h *TradeStatsHandler) Range(ctx context.Context, c *app.RequestContext) {
	data, err := h.statsSvc.GetDateRange()
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, data)
}

// Sync 从 summary 提取并刷新日终快照表（历史回填 / 手动补数）。
func (h *TradeStatsHandler) Sync(ctx context.Context, c *app.RequestContext) {
	var req service.SyncDailyRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	data, err := h.statsSvc.SyncDailySnapshots(&req)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, data)
}

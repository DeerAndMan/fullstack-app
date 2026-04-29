package handler

import (
	"context"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type TradeHandler struct {
	tradeSvc *service.TradeService
}

func NewTradeHandler(tradeSvc *service.TradeService) *TradeHandler {
	return &TradeHandler{tradeSvc: tradeSvc}
}

func (h *TradeHandler) Index(ctx context.Context, c *app.RequestContext) {
	var req service.TradeRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	list, err := h.tradeSvc.Index(&req)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, list)
}

func (h *TradeHandler) Summary(ctx context.Context, c *app.RequestContext) {
	var req service.TradeRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	data, err := h.tradeSvc.Summary(&req)
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

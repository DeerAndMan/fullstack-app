package v1

import (
	"context"
	"errors"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type MarketHandler struct {
	market *service.MarketService
}

func NewMarketHandler(market *service.MarketService) *MarketHandler {
	return &MarketHandler{market: market}
}

func (h *MarketHandler) Feed(ctx context.Context, c *app.RequestContext) {
	data, err := h.market.Feed(ctx, c.Param("feed"), c.Query("symbol"), c.Query("date"))
	writeMarketResult(ctx, c, data, err)
}

func (h *MarketHandler) ConceptReason(ctx context.Context, c *app.RequestContext) {
	data, err := h.market.ConceptReason(ctx, c.Param("concept"), c.Param("symbol"))
	writeMarketResult(ctx, c, data, err)
}

func writeMarketResult(ctx context.Context, c *app.RequestContext, data any, err error) {
	if errors.Is(err, service.ErrMarketQuery) {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, data)
}

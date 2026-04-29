package handler

import (
	"context"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type JyDataHandler struct {
	jyDataSvc *service.JyDataService
}

func NewJyDataHandler(jyDataSvc *service.JyDataService) *JyDataHandler {
	return &JyDataHandler{jyDataSvc: jyDataSvc}
}

func (h *JyDataHandler) GetLatest(ctx context.Context, c *app.RequestContext) {
	list, err := h.jyDataSvc.GetLatest()
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, map[string]interface{}{
		"total": len(list),
		"list":  list,
	})
}

func (h *JyDataHandler) ListByDateRange(ctx context.Context, c *app.RequestContext) {
	var req service.JyDataListRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	list, err := h.jyDataSvc.ListByDateRange(&req)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, list)
}

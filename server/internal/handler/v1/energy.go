package v1

import (
	"context"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type EnergyHandler struct {
	energySvc *service.EnergyService
}

func NewEnergyHandler(energySvc *service.EnergyService) *EnergyHandler {
	return &EnergyHandler{energySvc: energySvc}
}

func (h *EnergyHandler) InsertAssets(ctx context.Context, c *app.RequestContext) {
	var data []model.Assets
	if err := c.BindAndValidate(&data); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.energySvc.InsertAssets(data); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, data)
}

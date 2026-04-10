package handler

import (
	"context"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type UploadHandler struct {
	uploadSvc *service.UploadService
}

func NewUploadHandler(uploadSvc *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadSvc: uploadSvc}
}

func (h *UploadHandler) Upload(ctx context.Context, c *app.RequestContext) {
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, "file is required")
		return
	}

	info, err := h.uploadSvc.Upload(file)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, info)
}

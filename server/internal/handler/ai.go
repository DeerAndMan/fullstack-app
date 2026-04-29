package handler

import (
	"context"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type AiHandler struct {
	aiSvc *service.AiService
}

func NewAiHandler(aiSvc *service.AiService) *AiHandler {
	return &AiHandler{aiSvc: aiSvc}
}

func (h *AiHandler) GetConversations(ctx context.Context, c *app.RequestContext) {
	user := c.Query("user")
	if user == "" {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, "user is required")
		return
	}

	data, err := h.aiSvc.GetConversations(user)
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

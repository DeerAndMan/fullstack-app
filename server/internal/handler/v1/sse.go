package v1

import (
	"context"
	"io"
	"net/http"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type SseHandler struct {
	sseSvc *service.SseService
}

func NewSseHandler(sseSvc *service.SseService) *SseHandler {
	return &SseHandler{sseSvc: sseSvc}
}

func (h *SseHandler) ChatMessages(ctx context.Context, c *app.RequestContext) {
	var req service.ChatRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	resp, err := h.sseSvc.ChatMessages(&req)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	defer resp.Body.Close()

	c.SetStatusCode(http.StatusOK)
	c.Response.Header.Set("Content-Type", "application/json")
	buf := make([]byte, 1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			c.Write(buf[:n])
			c.Flush()
		}
		if readErr != nil {
			if readErr != io.EOF {
				response.Fail(ctx, c, errcode.ErrInternal)
			}
			break
		}
	}
}

func (h *SseHandler) GetChatMessages(ctx context.Context, c *app.RequestContext) {
	conversationID := c.Param("id")
	user := c.Query("user")
	if user == "" {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, "user is required")
		return
	}

	data, err := h.sseSvc.GetChatMessages(conversationID, user)
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

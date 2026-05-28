package v2

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type TestHandler struct{}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

func (h *TestHandler) Ping(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, map[string]any{
		"code":    0,
		"data":    "pong from v2",
		"message": "success",
	})
}

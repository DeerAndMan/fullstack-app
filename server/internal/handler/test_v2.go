package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type TestHandlerV2 struct{}

func NewTestHandlerV2() *TestHandlerV2 {
	return &TestHandlerV2{}
}

func (h *TestHandlerV2) Ping(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, map[string]any{
		"code":    0,
		"data":    "pong from v2",
		"message": "success",
	})
}

package response

import (
	"context"
	"net/http"

	"fullstack-app/server/pkg/errcode"

	"github.com/cloudwego/hertz/pkg/app"
)

type Body struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

func OK(ctx context.Context, c *app.RequestContext, data any) {
	c.JSON(http.StatusOK, Body{
		Code:    0,
		Data:    data,
		Message: "success",
	})
}

func OKWithMessage(ctx context.Context, c *app.RequestContext, msg string) {
	c.JSON(http.StatusOK, Body{
		Code:    0,
		Data:    nil,
		Message: msg,
	})
}

func Fail(ctx context.Context, c *app.RequestContext, err *errcode.Error) {
	c.JSON(err.HTTP, Body{
		Code:    err.Code,
		Data:    nil,
		Message: err.Message,
	})
}

func FailWithMessage(ctx context.Context, c *app.RequestContext, err *errcode.Error, msg string) {
	c.JSON(err.HTTP, Body{
		Code:    err.Code,
		Data:    nil,
		Message: msg,
	})
}

type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

func OKWithPage(ctx context.Context, c *app.RequestContext, list any, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Body{
		Code: 0,
		Data: PageResult{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		Message: "success",
	})
}

package v1

import (
	"context"
	"strconv"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type SubscriptionHandler struct {
	subSvc *service.SubscriptionService
}

func NewSubscriptionHandler(subSvc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subSvc: subSvc}
}

func (h *SubscriptionHandler) Create(ctx context.Context, c *app.RequestContext) {
	var sub model.XqSubscription
	if err := c.BindAndValidate(&sub); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	result, err := h.subSvc.Create(&sub)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, result)
}

func (h *SubscriptionHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	if err := h.subSvc.Delete(id, userID); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OKWithMessage(ctx, c, "订阅删除成功")
}

func (h *SubscriptionHandler) GetByID(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	sub, err := h.subSvc.GetByID(id, userID)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, sub)
}

func (h *SubscriptionHandler) GetByUserID(ctx context.Context, c *app.RequestContext) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	subs, err := h.subSvc.GetByUserID(userID)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, subs)
}

func (h *SubscriptionHandler) GetAll(ctx context.Context, c *app.RequestContext) {
	subs, err := h.subSvc.GetAll()
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, subs)
}

func (h *SubscriptionHandler) ToggleEnabled(ctx context.Context, c *app.RequestContext) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	var req service.ToggleEnabledRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.subSvc.ToggleEnabled(userID, &req); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OKWithMessage(ctx, c, "订阅状态更新成功")
}

func (h *SubscriptionHandler) Exists(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	exists, err := h.subSvc.Exists(id, userID)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, map[string]bool{"exists": exists})
}

func (h *SubscriptionHandler) AppendDescription(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	var req service.AppendDescriptionRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.subSvc.AppendDescription(id, userID, &req); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OKWithMessage(ctx, c, "曾用名添加成功")
}

func (h *SubscriptionHandler) Detail(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	detail, err := h.subSvc.Detail(id, userID)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, detail)
}

func (h *SubscriptionHandler) DetailTable(ctx context.Context, c *app.RequestContext) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	pageNumber, _ := strconv.Atoi(c.DefaultQuery("pageNumber", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := h.subSvc.DetailTable(userID, pageNumber, pageSize)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, result)
}

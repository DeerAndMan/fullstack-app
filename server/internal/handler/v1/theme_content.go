package v1

import (
	"context"
	"strconv"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type ThemeContentHandler struct {
	tcSvc *service.ThemeContentService
}

func NewThemeContentHandler(tcSvc *service.ThemeContentService) *ThemeContentHandler {
	return &ThemeContentHandler{tcSvc: tcSvc}
}

func (h *ThemeContentHandler) BatchCreate(ctx context.Context, c *app.RequestContext) {
	var req service.ThemeContentBatchRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	count, err := h.tcSvc.BatchCreate(&req)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, map[string]int{"count": count})
}

func (h *ThemeContentHandler) Create(ctx context.Context, c *app.RequestContext) {
	var req service.CreateThemeContentRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	tc, err := h.tcSvc.Create(&req)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, tc)
}

func (h *ThemeContentHandler) Update(ctx context.Context, c *app.RequestContext) {
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

	var req service.UpdateThemeContentRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	tc, err := h.tcSvc.Update(id, userID, &req)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, tc)
}

func (h *ThemeContentHandler) Delete(ctx context.Context, c *app.RequestContext) {
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

	if err := h.tcSvc.Delete(id, userID); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OKWithMessage(ctx, c, "删除成功")
}

func (h *ThemeContentHandler) GetByID(ctx context.Context, c *app.RequestContext) {
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

	tc, err := h.tcSvc.GetByID(id, userID)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, tc)
}

func (h *ThemeContentHandler) GetByUserID(ctx context.Context, c *app.RequestContext) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.tcSvc.GetByUserID(userID, limit, offset)
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

func (h *ThemeContentHandler) GetAll(ctx context.Context, c *app.RequestContext) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.tcSvc.GetAll(limit, offset)
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

func (h *ThemeContentHandler) Exists(ctx context.Context, c *app.RequestContext) {
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

	exists, err := h.tcSvc.Exists(id, userID)
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

func (h *ThemeContentHandler) Search(ctx context.Context, c *app.RequestContext) {
	keyword := c.DefaultQuery("keyword", "")
	userID, _ := strconv.ParseInt(c.DefaultQuery("user_id", "0"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.tcSvc.Search(keyword, userID, limit, offset)
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

func (h *ThemeContentHandler) SaveTimeline(ctx context.Context, c *app.RequestContext) {
	var req service.SaveTimelineRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.tcSvc.SaveTimeline(&req); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OKWithMessage(ctx, c, "保存成功")
}

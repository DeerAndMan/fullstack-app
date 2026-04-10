package handler

import (
	"context"
	"strconv"

	"fullstack-app/server/internal/middleware"
	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) Create(ctx context.Context, c *app.RequestContext) {
	var req service.CreateUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.userSvc.Create(&req); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithMessage(ctx, c, "created")
}

func (h *UserHandler) GetByID(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	user, err := h.userSvc.GetByID(uint(id))
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, user)
}

func (h *UserHandler) GetProfile(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserID(c)
	user, err := h.userSvc.GetByID(userID)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, user)
}

func (h *UserHandler) Update(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	var req service.UpdateUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.userSvc.Update(uint(id), &req); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithMessage(ctx, c, "updated")
}

func (h *UserHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	if err := h.userSvc.Delete(uint(id)); err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithMessage(ctx, c, "deleted")
}

func (h *UserHandler) List(ctx context.Context, c *app.RequestContext) {
	var req service.ListUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	users, total, err := h.userSvc.List(&req)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithPage(ctx, c, users, total, req.Page, req.PageSize)
}

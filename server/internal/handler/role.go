package handler

import (
	"context"
	"strconv"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type RoleHandler struct {
	roleSvc *service.RoleService
}

func NewRoleHandler(roleSvc *service.RoleService) *RoleHandler {
	return &RoleHandler{roleSvc: roleSvc}
}

func (h *RoleHandler) Create(ctx context.Context, c *app.RequestContext) {
	var req service.CreateRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.roleSvc.Create(&req); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithMessage(ctx, c, "created")
}

func (h *RoleHandler) GetByID(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	role, err := h.roleSvc.GetByID(uint(id))
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, role)
}

func (h *RoleHandler) Update(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	var req service.UpdateRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.roleSvc.Update(uint(id), &req); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithMessage(ctx, c, "updated")
}

func (h *RoleHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	if err := h.roleSvc.Delete(uint(id)); err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithMessage(ctx, c, "deleted")
}

func (h *RoleHandler) List(ctx context.Context, c *app.RequestContext) {
	var req service.ListRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	roles, total, err := h.roleSvc.List(&req)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithPage(ctx, c, roles, total, req.Page, req.PageSize)
}

func (h *RoleHandler) GetAll(ctx context.Context, c *app.RequestContext) {
	roles, err := h.roleSvc.GetAll()
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, roles)
}

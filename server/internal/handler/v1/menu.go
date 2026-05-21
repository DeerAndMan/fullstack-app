package v1

import (
	"context"
	"strconv"

	"fullstack-app/server/internal/middleware"
	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type MenuHandler struct {
	menuSvc *service.MenuService
}

func NewMenuHandler(menuSvc *service.MenuService) *MenuHandler {
	return &MenuHandler{menuSvc: menuSvc}
}

func (h *MenuHandler) List(ctx context.Context, c *app.RequestContext) {
	menus, err := h.menuSvc.List()
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, menus)
}

func (h *MenuHandler) AddAll(ctx context.Context, c *app.RequestContext) {
	var req []service.AddMenuRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	count, err := h.menuSvc.AddAll(req)
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

func (h *MenuHandler) RoleBinding(ctx context.Context, c *app.RequestContext) {
	var req service.RoleBindingRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	currentUserID := int64(middleware.GetUserID(c))
	if err := h.menuSvc.RoleBinding(&req, currentUserID); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OKWithMessage(ctx, c, "绑定成功")
}

func (h *MenuHandler) ListByRoleID(ctx context.Context, c *app.RequestContext) {
	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 64)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}

	menus, err := h.menuSvc.ListByRoleID(roleID)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, menus)
}

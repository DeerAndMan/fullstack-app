package v1

import (
	"context"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type EnumHandler struct {
	roleSvc *service.RoleService
}

func NewEnumHandler(roleSvc *service.RoleService) *EnumHandler {
	return &EnumHandler{roleSvc: roleSvc}
}

type EnumRole struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	RoleKey string `json:"role_key"`
}

func (h *EnumHandler) Roles(ctx context.Context, c *app.RequestContext) {
	roles, err := h.roleSvc.GetAll()
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	list := make([]EnumRole, 0, len(roles))
	for _, r := range roles {
		list = append(list, EnumRole{
			ID:      r.RoleID,
			Name:    r.RoleName,
			RoleKey: r.RoleKey,
		})
	}
	response.OK(ctx, c, list)
}

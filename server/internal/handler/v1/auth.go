package v1

import (
	"context"
	"strings"

	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req service.RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.authSvc.Register(&req); err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithMessage(ctx, c, "register success")
}

func (h *AuthHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req service.LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	loginResp, err := h.authSvc.Login(&req)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, loginResp)
}

func (h *AuthHandler) RefreshToken(ctx context.Context, c *app.RequestContext) {
	var req service.RefreshTokenRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, err.Error())
		return
	}

	tokenPair, err := h.authSvc.RefreshToken(&req)
	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Fail(ctx, c, e)
			return
		}
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OK(ctx, c, tokenPair)
}

func (h *AuthHandler) Logout(ctx context.Context, c *app.RequestContext) {
	// 从 Authorization 头取出 access 令牌用于吊销
	auth := string(c.GetHeader("Authorization"))
	accessToken := strings.TrimPrefix(auth, "Bearer ")

	if err := h.authSvc.Logout(accessToken); err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}

	response.OKWithMessage(ctx, c, "logout success")
}

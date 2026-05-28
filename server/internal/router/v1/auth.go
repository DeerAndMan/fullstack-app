package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerAuthRoutes(public *route.RouterGroup, protected *route.RouterGroup, h *handlerv1.AuthHandler) {
	auth := public.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh-token", h.RefreshToken)
	}

	protected.POST("/auth/logout", h.Logout)
}

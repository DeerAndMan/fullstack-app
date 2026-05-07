package v2

import (
	"fullstack-app/server/internal/handler"
	"fullstack-app/server/internal/middleware"
	jwtpkg "fullstack-app/server/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"
)

type Handlers struct {
	Test *handler.TestHandlerV2
}

func (h *Handlers) Register(srv *server.Hertz, jwtManager *jwtpkg.Manager) {
	apiV2 := srv.Group("/api/v2")
	protectedV2 := apiV2.Group("")
	protectedV2.Use(middleware.JWTAuth(jwtManager))
	RegisterRoutes(apiV2, protectedV2, h)
}

func RegisterRoutes(public *route.RouterGroup, protected *route.RouterGroup, h *Handlers) {
	registerTestRoutes(public, h.Test)
}

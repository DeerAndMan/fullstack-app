package router

import (
	"fullstack-app/server/internal/middleware"
	v1 "fullstack-app/server/internal/router/v1"
	v2 "fullstack-app/server/internal/router/v2"
	jwtpkg "fullstack-app/server/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Setup(h *server.Hertz, jwtManager *jwtpkg.Manager, allowOrigins []string, v1Handlers *v1.Handlers, v2Handlers *v2.Handlers) {
	h.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.CORS(allowOrigins),
		middleware.Logger(),
	)

	// 公共路由(无版本前缀)
	RegisterPublicRoutes(h, v1Handlers)

	v1Handlers.Register(h, jwtManager)
	v2Handlers.Register(h, jwtManager)
}

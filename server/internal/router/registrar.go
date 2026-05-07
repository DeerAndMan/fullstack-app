package router

import (
	jwtpkg "fullstack-app/server/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app/server"
)

type Registrar interface {
	Register(h *server.Hertz, jwtManager *jwtpkg.Manager)
}

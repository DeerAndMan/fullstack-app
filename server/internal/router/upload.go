package router

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerUploadRoutes(protected *route.RouterGroup, h *handler.UploadHandler) {
	protected.POST("/upload", h.Upload)
}

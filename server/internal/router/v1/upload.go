package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerUploadRoutes(protected *route.RouterGroup, h *handlerv1.UploadHandler) {
	protected.POST("/upload", h.Upload)
}

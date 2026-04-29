package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerJyDataRoutes(public *route.RouterGroup, h *handler.JyDataHandler) {
	api := public.Group("/jydata")
	{
		api.GET("/latest", h.GetLatest)
		api.POST("/list", h.ListByDateRange)
	}
}

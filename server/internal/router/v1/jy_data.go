package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerJyDataRoutes(public *route.RouterGroup, h *handlerv1.JyDataHandler) {
	api := public.Group("/jydata")
	{
		api.GET("/latest", h.GetLatest)
		api.POST("/list", h.ListByDateRange)
	}
}

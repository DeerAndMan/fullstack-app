package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerEnergyRoutes(public *route.RouterGroup, h *handlerv1.EnergyHandler) {
	energy := public.Group("/energy")
	{
		energy.POST("/insert", h.InsertAssets)
		energy.POST("/asset", h.InsertAssets)
	}
}

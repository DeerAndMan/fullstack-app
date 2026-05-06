package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerEnergyRoutes(public *route.RouterGroup, h *handler.EnergyHandler) {
	energy := public.Group("/energy")
	{
		energy.POST("/insert", h.InsertAssets)
		energy.POST("/asset", h.InsertAssets)
	}
}

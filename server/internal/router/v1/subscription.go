package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerSubscriptionRoutes(public *route.RouterGroup, h *handlerv1.SubscriptionHandler) {
	subs := public.Group("/subscriptions")
	{
		subs.POST("", h.Create)
		subs.GET("", h.GetAll)
		subs.DELETE("/:id/:userId", h.Delete)
		subs.GET("/:id/:userId", h.GetByID)
		subs.GET("/user/:userId", h.GetByUserID)
		subs.PUT("/toggle/:userId", h.ToggleEnabled)
		subs.GET("/exists/:id/:userId", h.Exists)
		subs.PUT("/description/:id/:userId", h.AppendDescription)
		subs.GET("/detail/:id/:userId", h.Detail)
		subs.GET("/detail-table/:userId", h.DetailTable)
	}
}

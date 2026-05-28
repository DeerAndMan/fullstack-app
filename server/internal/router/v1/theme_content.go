package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerThemeContentRoutes(public *route.RouterGroup, h *handlerv1.ThemeContentHandler) {
	tc := public.Group("/theme-contents")
	{
		tc.POST("/batch", h.BatchCreate)
		tc.POST("", h.Create)
		tc.GET("", h.GetAll)
		tc.PUT("/:id/:userId", h.Update)
		tc.DELETE("/:id/:userId", h.Delete)
		tc.GET("/:id/:userId", h.GetByID)
		tc.GET("/user/:userId", h.GetByUserID)
		tc.GET("/exists/:id/:userId", h.Exists)
		tc.GET("/search", h.Search)
		tc.POST("/timeline", h.SaveTimeline)
	}
}

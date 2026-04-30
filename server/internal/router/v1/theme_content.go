package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerThemeContentRoutes(public *route.RouterGroup, h *handler.ThemeContentHandler) {
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

package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"

	"github.com/cloudwego/hertz/pkg/route"
)

func registerTopicResearchRoutes(protected *route.RouterGroup, handler *handlerv1.TopicResearchHandler) {
	research := protected.Group("/topic-research")
	research.GET("/skill-catalog", handler.SkillCatalog)
	research.POST("/skill-groups", handler.SaveSkillGroup)
	research.DELETE("/skill-groups/:id", handler.DeleteSkillGroup)
	research.GET("/skills", handler.Skills)
	research.PUT("/skills/order", handler.OrderSkills)
	research.POST("/skills", handler.SaveSkill)
	research.GET("/runs", handler.Runs)
	research.GET("/runs/:id", handler.Run)
	research.POST("/runs", handler.Start)
}

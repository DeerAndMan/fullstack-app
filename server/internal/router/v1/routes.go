package v1

import (
	handlerv1 "fullstack-app/server/internal/handler/v1"
	"fullstack-app/server/internal/middleware"
	jwtpkg "fullstack-app/server/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"
)

type Handlers struct {
	Auth         *handlerv1.AuthHandler
	User         *handlerv1.UserHandler
	Role         *handlerv1.RoleHandler
	Upload       *handlerv1.UploadHandler
	Energy       *handlerv1.EnergyHandler
	Trade        *handlerv1.TradeHandler
	TradeStats   *handlerv1.TradeStatsHandler
	JyData       *handlerv1.JyDataHandler
	Sse          *handlerv1.SseHandler
	Ai           *handlerv1.AiHandler
	Enum         *handlerv1.EnumHandler
	Menu         *handlerv1.MenuHandler
	Subscription *handlerv1.SubscriptionHandler
	ThemeContent *handlerv1.ThemeContentHandler
	Market       *handlerv1.MarketHandler
	Review       *handlerv1.MarketReviewHandler
	Research     *handlerv1.TopicResearchHandler
}

func (h *Handlers) Register(srv *server.Hertz, jwtManager *jwtpkg.Manager) {
	apiV1 := srv.Group("/api/v1")
	protectedV1 := apiV1.Group("")                  // 受保护组
	protectedV1.Use(middleware.JWTAuth(jwtManager)) // 添加 JWT 验证中间件
	apiV1.GET("/topic-research/stream", h.Research.Stream(jwtManager))
	RegisterRoutes(apiV1, protectedV1, h)
}

func RegisterRoutes(public *route.RouterGroup, protected *route.RouterGroup, h *Handlers) {
	registerAuthRoutes(public, protected, h.Auth)
	registerUserRoutes(protected, h.User)
	registerRoleRoutes(protected, h.Role)
	registerUploadRoutes(protected, h.Upload)
	registerEnergyRoutes(public, h.Energy)
	registerTradeRoutes(protected, h.Trade)
	registerTradeStatsRoutes(protected, h.TradeStats)
	protected.GET("/review/dates", h.Review.Dates)
	protected.GET("/review/refresh", h.Review.RefreshStatus)
	protected.POST("/review/refresh", h.Review.StartRefresh)
	protected.POST("/review/advice", h.Research.ReviewAdvice)
	protected.GET("/review/timeline", h.Review.Timeline)
	protected.GET("/review/daily", h.Review.Daily)
	protected.GET("/review/members", h.Review.Members)
	protected.GET("/review/study", h.Review.Study)
	protected.GET("/review/note", h.Review.Note)
	protected.PUT("/review/note", h.Review.SaveNote)
	protected.GET("/market/:feed", h.Market.Feed)
	protected.GET("/market/concepts/:concept/stocks/:symbol/reason", h.Market.ConceptReason)
	registerTopicResearchRoutes(protected, h.Research)
	registerJyDataRoutes(public, h.JyData)
	registerSseRoutes(public, h.Sse)
	registerAiRoutes(public, h.Ai)
	registerEnumRoutes(public, h.Enum)
	registerMenuRoutes(protected, h.Menu)
	registerSubscriptionRoutes(public, h.Subscription)
	registerThemeContentRoutes(public, h.ThemeContent)
}

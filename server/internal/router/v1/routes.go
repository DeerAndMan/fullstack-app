package v1

import (
	"fullstack-app/server/internal/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

type Handlers struct {
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	Role         *handler.RoleHandler
	Upload       *handler.UploadHandler
	Energy       *handler.EnergyHandler
	Trade        *handler.TradeHandler
	JyData       *handler.JyDataHandler
	Sse          *handler.SseHandler
	Ai           *handler.AiHandler
	Enum         *handler.EnumHandler
	Menu         *handler.MenuHandler
	Subscription *handler.SubscriptionHandler
	ThemeContent *handler.ThemeContentHandler
}

func RegisterRoutes(public *route.RouterGroup, protected *route.RouterGroup, h *Handlers) {
	registerAuthRoutes(public, protected, h.Auth)
	registerUserRoutes(protected, h.User)
	registerRoleRoutes(protected, h.Role)
	registerUploadRoutes(protected, h.Upload)
	registerEnergyRoutes(public, h.Energy)
	registerTradeRoutes(protected, h.Trade)
	registerJyDataRoutes(public, h.JyData)
	registerSseRoutes(public, h.Sse)
	registerAiRoutes(public, h.Ai)
	registerEnumRoutes(public, h.Enum)
	registerMenuRoutes(protected, h.Menu)
	registerSubscriptionRoutes(public, h.Subscription)
	registerThemeContentRoutes(public, h.ThemeContent)
}

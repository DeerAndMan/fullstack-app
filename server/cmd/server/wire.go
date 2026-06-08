package main

import (
	"fullstack-app/server/internal/config"
	handlerv1 "fullstack-app/server/internal/handler/v1"
	handlerv2 "fullstack-app/server/internal/handler/v2"
	"fullstack-app/server/internal/repository"
	v1 "fullstack-app/server/internal/router/v1"
	v2 "fullstack-app/server/internal/router/v2"
	"fullstack-app/server/internal/service"
	jwtpkg "fullstack-app/server/pkg/jwt"
	"fullstack-app/server/pkg/upload"

	"gorm.io/gorm"
)

// AppDeps 汇总了 main 中需要直接使用的依赖（例如 WsHub 需要在外部触发广播）。
type AppDeps struct {
	V1    *v1.Handlers
	V2    *v2.Handlers
	WsHub *service.WsHub
}

func initHandlers(db *gorm.DB, jwtManager *jwtpkg.Manager, uploader *upload.Uploader, cfg *config.Config) *AppDeps {
	// Repository
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	energyRepo := repository.NewEnergyRepository(db)
	jyDataRepo := repository.NewJyDataRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	subRepo := repository.NewSubscriptionRepository(db)
	tcRepo := repository.NewThemeContentRepository(db)

	// Service
	authSvc := service.NewAuthService(userRepo, jwtManager)
	userSvc := service.NewUserService(userRepo, roleRepo)
	roleSvc := service.NewRoleService(roleRepo)
	uploadSvc := service.NewUploadService(uploader)
	energySvc := service.NewEnergyService(energyRepo)
	tradeSvc := service.NewTradeService(energyRepo)
	jyDataSvc := service.NewJyDataService(jyDataRepo)
	sseSvc := service.NewSseService(cfg.AI.BaseURL, cfg.AI.Token)
	aiSvc := service.NewAiService(cfg.AI.BaseURL, cfg.AI.Token)
	menuSvc := service.NewMenuService(menuRepo)
	subSvc := service.NewSubscriptionService(subRepo, tcRepo)
	tcSvc := service.NewThemeContentService(tcRepo, subSvc)
	wsHub := service.NewWsHub()

	// v1 Handlers
	handlers := &v1.Handlers{
		Auth:         handlerv1.NewAuthHandler(authSvc),
		User:         handlerv1.NewUserHandler(userSvc),
		Role:         handlerv1.NewRoleHandler(roleSvc),
		Upload:       handlerv1.NewUploadHandler(uploadSvc),
		Energy:       handlerv1.NewEnergyHandler(energySvc),
		Trade:        handlerv1.NewTradeHandler(tradeSvc),
		JyData:       handlerv1.NewJyDataHandler(jyDataSvc),
		Sse:          handlerv1.NewSseHandler(sseSvc),
		Ai:           handlerv1.NewAiHandler(aiSvc),
		Enum:         handlerv1.NewEnumHandler(roleSvc),
		Menu:         handlerv1.NewMenuHandler(menuSvc),
		Subscription: handlerv1.NewSubscriptionHandler(subSvc),
		ThemeContent: handlerv1.NewThemeContentHandler(tcSvc),
	}

	// v2 Handlers
	handlersV2 := &v2.Handlers{
		Test: handlerv2.NewTestHandler(),
		Ws:   handlerv2.NewWsHandler(wsHub),
	}

	return &AppDeps{V1: handlers, V2: handlersV2, WsHub: wsHub}
}

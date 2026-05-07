package main

import (
	"fullstack-app/server/internal/config"
	"fullstack-app/server/internal/handler"
	"fullstack-app/server/internal/repository"
	v1 "fullstack-app/server/internal/router/v1"
	v2 "fullstack-app/server/internal/router/v2"
	"fullstack-app/server/internal/service"
	jwtpkg "fullstack-app/server/pkg/jwt"
	"fullstack-app/server/pkg/upload"

	"gorm.io/gorm"
)

func initHandlers(db *gorm.DB, jwtManager *jwtpkg.Manager, uploader *upload.Uploader, cfg *config.Config) (*v1.Handlers, *v2.Handlers) {
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

	// v1 Handlers
	handlers := &v1.Handlers{
		Auth:         handler.NewAuthHandler(authSvc),
		User:         handler.NewUserHandler(userSvc),
		Role:         handler.NewRoleHandler(roleSvc),
		Upload:       handler.NewUploadHandler(uploadSvc),
		Energy:       handler.NewEnergyHandler(energySvc),
		Trade:        handler.NewTradeHandler(tradeSvc),
		JyData:       handler.NewJyDataHandler(jyDataSvc),
		Sse:          handler.NewSseHandler(sseSvc),
		Ai:           handler.NewAiHandler(aiSvc),
		Enum:         handler.NewEnumHandler(roleSvc),
		Menu:         handler.NewMenuHandler(menuSvc),
		Subscription: handler.NewSubscriptionHandler(subSvc),
		ThemeContent: handler.NewThemeContentHandler(tcSvc),
	}

	// v2 Handlers
	handlersV2 := &v2.Handlers{
		Test: handler.NewTestHandlerV2(),
	}

	return handlers, handlersV2
}

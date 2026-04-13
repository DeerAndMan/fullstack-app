package main

import (
	"flag"
	"fmt"

	"fullstack-app/server/internal/config"
	"fullstack-app/server/internal/database"
	"fullstack-app/server/internal/handler"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/internal/router"
	"fullstack-app/server/internal/service"
	jwtpkg "fullstack-app/server/pkg/jwt"
	"fullstack-app/server/pkg/upload"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "config file path")
	flag.Parse()

	// Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// Config
	cfg, err := config.Load(*configPath)
	if err != nil {
		zap.L().Fatal("load config failed", zap.Error(err))
	}

	// MySQL
	db, err := database.NewMySQL(&cfg.MySQL)
	if err != nil {
		zap.L().Fatal("connect mysql failed", zap.Error(err))
	}
	if err := database.AutoMigrate(db); err != nil {
		zap.L().Fatal("auto migrate failed", zap.Error(err))
	}

	// Redis
	_, err = database.NewRedis(&cfg.Redis)
	if err != nil {
		zap.L().Fatal("connect redis failed", zap.Error(err))
	}

	// JWT
	jwtManager := jwtpkg.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire,
		cfg.JWT.Issuer,
	)

	// Uploader
	uploader := upload.NewUploader(cfg.Upload.Path, cfg.Upload.MaxSize, cfg.Upload.AllowExts)

	// Repository
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	// Service
	authSvc := service.NewAuthService(userRepo, jwtManager)
	userSvc := service.NewUserService(userRepo)
	roleSvc := service.NewRoleService(roleRepo)
	uploadSvc := service.NewUploadService(uploader)

	// Handler
	handlers := &router.Handlers{
		Auth:   handler.NewAuthHandler(authSvc),
		User:   handler.NewUserHandler(userSvc),
		Role:   handler.NewRoleHandler(roleSvc),
		Upload: handler.NewUploadHandler(uploadSvc),
	}

	// Hertz server
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("0.0.0.0:%d", cfg.Server.Port)),
		server.WithMaxRequestBodySize(20*1024*1024), // 20MB
	)

	// Routes
	router.Setup(h, handlers, jwtManager)

	zap.L().Info("server starting", zap.Int("port", cfg.Server.Port))
	h.Spin()
}

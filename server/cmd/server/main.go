package main

import (
	"flag"
	"fmt"
	"os"

	"fullstack-app/server/internal/config"
	"fullstack-app/server/internal/database"
	"fullstack-app/server/internal/handler"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/internal/router"
	"fullstack-app/server/internal/service"
	jwtpkg "fullstack-app/server/pkg/jwt"
	"fullstack-app/server/pkg/upload"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "config file path")
	flag.Parse()

	// Logger (fatih/color 彩色控制台输出)
	colorLevel := func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		var s string
		switch l {
		case zapcore.DebugLevel:
			s = color.CyanString("DEBUG")
		case zapcore.InfoLevel:
			s = color.GreenString("INFO")
		case zapcore.WarnLevel:
			s = color.YellowString("WARN")
		case zapcore.ErrorLevel:
			s = color.RedString("ERROR")
		case zapcore.FatalLevel:
			s = color.New(color.FgRed, color.Bold).Sprint("FATAL")
		default:
			s = l.CapitalString()
		}
		enc.AppendString(s)
	}
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = colorLevel
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05")
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	), zap.AddCaller())
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// Config
	cfg, err := config.Load(*configPath)
	if err != nil {
		zap.L().Fatal("load config failed", zap.Error(err))
	}

	// MySQL
	db, err := database.NewMySQL(&cfg.MySQL, cfg.Server.Mode)
	if err != nil {
		zap.L().Fatal("connect mysql failed", zap.Error(err))
	}
	if err := database.AutoMigrate(db); err != nil {
		zap.L().Fatal("auto migrate failed", zap.Error(err))
	}

	// Redis (应用层暂未使用，连接失败不阻塞启动)
	_, err = database.NewRedis(&cfg.Redis)
	if err != nil {
		zap.L().Warn("connect redis failed, skipping", zap.Error(err))
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
		User:   handler.NewUserHandler(userSvc, uploadSvc),
		Role:   handler.NewRoleHandler(roleSvc),
		Upload: handler.NewUploadHandler(uploadSvc),
	}

	// Hertz server
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("0.0.0.0:%d", cfg.Server.Port)),
		server.WithMaxRequestBodySize(20*1024*1024), // 20MB
	)

	// Routes
	router.Setup(h, handlers, jwtManager, cfg.CORS.AllowOrigins)

	// Startup banner
	fmt.Println()
	fmt.Println(color.CyanString("  ┌─────────────────────────────────────────┐"))
	fmt.Println(color.CyanString("  │") + color.GreenString("   Fullstack App Server Ready              ") + color.CyanString("│"))
	fmt.Println(color.CyanString("  ├─────────────────────────────────────────┤"))
	fmt.Printf("  %s  %-8s %s\n", color.CyanString("│"), color.YellowString("Local:"), color.GreenString("http://localhost:%d", cfg.Server.Port))
	fmt.Printf("  %s  %-8s %s\n", color.CyanString("│"), color.YellowString("Mode:"), cfg.Server.Mode)
	fmt.Printf("  %s  %-8s %s\n", color.CyanString("│"), color.YellowString("MySQL:"), fmt.Sprintf("%s:%d/%s", cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database))
	fmt.Printf("  %s  %-8s %s\n", color.CyanString("│"), color.YellowString("Redis:"), cfg.Redis.Addr())
	fmt.Println(color.CyanString("  └─────────────────────────────────────────┘"))
	fmt.Println()

	h.Spin()
}

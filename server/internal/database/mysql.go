package database

import (
	"fmt"
	"time"

	"fullstack-app/server/internal/config"
	"fullstack-app/server/internal/model"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQL(cfg *config.MySQLConfig, serverMode string) (*gorm.DB, error) {
	logLevel := logger.Info
	if serverMode == "release" {
		logLevel = logger.Warn
	}
	if cfg.Host == "" {
		return nil, fmt.Errorf("mysql host is required")
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logLevel),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	zap.L().Info("mysql connected", zap.String("host", cfg.Host), zap.String("database", cfg.Database))
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Role{},
	)
}

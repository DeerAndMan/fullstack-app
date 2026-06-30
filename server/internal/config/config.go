package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Upload UploadConfig `mapstructure:"upload"`
	CORS   CORSConfig   `mapstructure:"cors"`
	AI     AIConfig     `mapstructure:"ai"`
}

type AIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Token   string `mapstructure:"token"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug / release
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type JWTConfig struct {
	Secret        string  `mapstructure:"secret"`
	AccessExpire  float64 `mapstructure:"access_expire"`  // 单位：小时，支持小数（如 0.1 表示 6 分钟）
	RefreshExpire float64 `mapstructure:"refresh_expire"` // 单位：小时，支持小数
	Issuer        string  `mapstructure:"issuer"`
}

type UploadConfig struct {
	Path      string   `mapstructure:"path"`
	MaxSize   int64    `mapstructure:"max_size"` // MB
	AllowExts []string `mapstructure:"allow_exts"`
}

type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("APP")
	// 将嵌套 key 的点替换为下划线，使 APP_MYSQL_HOST 这类环境变量能覆盖 mysql.host
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	zap.L().Info("config loaded", zap.String("file", path))
	return &cfg, nil
}

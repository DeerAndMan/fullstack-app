package test

import (
	"os"
	"path/filepath"
	"testing"

	"fullstack-app/server/internal/config"
)

const testConfigYAML = `
server:
  port: 6767
  mode: release
mysql:
  host: db.example.test
  port: 3307
  username: tester
  password: secret
  database: app_test
  max_idle_conns: 3
  max_open_conns: 9
redis:
  host: redis.example.test
  port: 6380
  password: redis-secret
  db: 2
jwt:
  access_expire: 1.5
  refresh_expire: 72
upload:
  path: ./tmp/uploads
  max_size: 12
  allow_exts:
    - .jpg
    - .png
cors:
  allow_origins:
    - http://localhost:3000
ai:
  base_url: https://ai.example.test
  token: ai-secret
`

func writeTestConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(testConfigYAML), 0o600); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	return path
}

func TestConfigLoadAndDerivedValues(t *testing.T) {
	path := writeTestConfig(t)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 6767 || cfg.Server.Mode != "release" {
		t.Fatalf("unexpected server config: %+v", cfg.Server)
	}
	if cfg.MySQL.Host != "db.example.test" || cfg.MySQL.Port != 3307 || cfg.MySQL.Username != "tester" || cfg.MySQL.Password != "secret" || cfg.MySQL.Database != "app_test" {
		t.Fatalf("unexpected mysql config: %+v", cfg.MySQL)
	}
	if cfg.MySQL.MaxIdleConns != 3 || cfg.MySQL.MaxOpenConns != 9 {
		t.Fatalf("unexpected mysql pool config: %+v", cfg.MySQL)
	}
	if got, want := cfg.MySQL.DSN(), "tester:secret@tcp(db.example.test:3307)/app_test?charset=utf8mb4&parseTime=True&loc=Local"; got != want {
		t.Fatalf("MySQL.DSN() = %q, want %q", got, want)
	}
	if cfg.Redis.Host != "redis.example.test" || cfg.Redis.Port != 6380 || cfg.Redis.Password != "redis-secret" || cfg.Redis.DB != 2 {
		t.Fatalf("unexpected redis config: %+v", cfg.Redis)
	}
	if got, want := cfg.Redis.Addr(), "redis.example.test:6380"; got != want {
		t.Fatalf("Redis.Addr() = %q, want %q", got, want)
	}
	if cfg.JWT.AccessExpire != 1.5 || cfg.JWT.RefreshExpire != 72 {
		t.Fatalf("unexpected jwt config: %+v", cfg.JWT)
	}
	if cfg.Upload.Path != "./tmp/uploads" || cfg.Upload.MaxSize != 12 || len(cfg.Upload.AllowExts) != 2 || cfg.Upload.AllowExts[1] != ".png" {
		t.Fatalf("unexpected upload config: %+v", cfg.Upload)
	}
	if len(cfg.CORS.AllowOrigins) != 1 || cfg.CORS.AllowOrigins[0] != "http://localhost:3000" {
		t.Fatalf("unexpected cors config: %+v", cfg.CORS)
	}
	if cfg.AI.BaseURL != "https://ai.example.test" || cfg.AI.Token != "ai-secret" {
		t.Fatalf("unexpected ai config: %+v", cfg.AI)
	}
}

func TestConfigLoadEnvironmentOverrides(t *testing.T) {
	path := writeTestConfig(t)
	t.Setenv("APP_SERVER_PORT", "8088")
	t.Setenv("APP_REDIS_PORT", "6390")
	t.Setenv("APP_JWT_ACCESS_EXPIRE", "0.25")
	t.Setenv("APP_UPLOAD_MAX_SIZE", "20")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Port != 8088 {
		t.Errorf("server.port = %d, want 8088", cfg.Server.Port)
	}
	if cfg.Redis.Port != 6390 {
		t.Errorf("redis.port = %d, want 6390", cfg.Redis.Port)
	}
	if cfg.JWT.AccessExpire != 0.25 {
		t.Errorf("jwt.access_expire = %v, want 0.25", cfg.JWT.AccessExpire)
	}
	if cfg.Upload.MaxSize != 20 {
		t.Errorf("upload.max_size = %d, want 20", cfg.Upload.MaxSize)
	}
}

func TestConfigLoadMissingFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("Load() error = nil, want an error for a missing file")
	}
}

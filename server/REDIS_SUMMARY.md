# Redis 登录功能实现总结

## ✅ 已完成的修改

### 1. `/cmd/server/main.go`
**修改前**：
```go
// Redis (应用层暂未使用，连接失败不阻塞启动)
_, err = database.NewRedis(&cfg.Redis)
if err != nil {
    zap.L().Warn("connect redis failed, skipping", zap.Error(err))
}
```

**修改后**：
```go
// Redis
rdb, err := database.NewRedis(&cfg.Redis)
if err != nil {
    zap.L().Fatal("connect redis failed", zap.Error(err))
}
```

**变化**：
- Redis 连接失败现在会导致程序退出（Fatal）
- 保存 Redis 实例到 `rdb` 变量
- 传递给 `initHandlers(db, rdb, jwtManager, uploader, cfg)`

---

### 2. `/cmd/server/wire.go`
**修改前**：
```go
import (
    ...
    "gorm.io/gorm"
)

func initHandlers(db *gorm.DB, jwtManager *jwtpkg.Manager, uploader *upload.Uploader, cfg *config.Config) *AppDeps {
    ...
    authSvc := service.NewAuthService(userRepo, jwtManager)
    ...
}
```

**修改后**：
```go
import (
    ...
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

func initHandlers(db *gorm.DB, rdb *redis.Client, jwtManager *jwtpkg.Manager, uploader *upload.Uploader, cfg *config.Config) *AppDeps {
    ...
    authSvc := service.NewAuthService(userRepo, jwtManager, rdb)
    ...
}
```

**变化**：
- 新增 `redis.Client` 导入
- 函数签名新增 `rdb *redis.Client` 参数
- `AuthService` 初始化时传入 Redis 客户端

---

### 3. `/internal/service/auth.go`
**修改前**：
```go
package service

import (
    "errors"
    "strconv"
    ...
)

type AuthService struct {
    userRepo   *repository.UserRepository
    jwtManager *jwtpkg.Manager
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *jwtpkg.Manager) *AuthService {
    return &AuthService{userRepo: userRepo, jwtManager: jwtManager}
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
    // ... 登录逻辑 ...
    
    return &LoginResponse{
        Token:     tokenPair,
        User:      userResp,
        Role:      role,
        MenuRoles: menuRoles,
    }, nil
}
```

**修改后**：
```go
package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "strconv"
    "time"
    ...
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

type AuthService struct {
    userRepo   *repository.UserRepository
    jwtManager *jwtpkg.Manager
    rdb        *redis.Client  // 新增
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *jwtpkg.Manager, rdb *redis.Client) *AuthService {
    return &AuthService{
        userRepo:   userRepo,
        jwtManager: jwtManager,
        rdb:        rdb,  // 新增
    }
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
    // ... 原有登录逻辑 ...
    
    // 将登录信息写入 Redis（新增）
    if err := s.saveLoginInfoToRedis(user.ID, user.Name, tokenPair.AccessToken); err != nil {
        zap.L().Error("保存登录信息到 Redis 失败", zap.Error(err), zap.Uint("userID", user.ID))
        // 不阻塞登录流程，仅记录日志
    }

    return &LoginResponse{
        Token:     tokenPair,
        User:      userResp,
        Role:      role,
        MenuRoles: menuRoles,
    }, nil
}

// 新增方法
func (s *AuthService) saveLoginInfoToRedis(userID uint, username, accessToken string) error {
    ctx := context.Background()

    loginInfo := map[string]any{
        "user_id":      userID,
        "username":     username,
        "access_token": accessToken,
        "login_time":   time.Now().Format("2006-01-02 15:04:05"),
    }

    data, err := json.Marshal(loginInfo)
    if err != nil {
        return fmt.Errorf("序列化登录信息失败: %w", err)
    }

    key := fmt.Sprintf("user:login:%d", userID)
    if err := s.rdb.Set(ctx, key, data, 7*24*time.Hour).Err(); err != nil {
        return fmt.Errorf("写入 Redis 失败: %w", err)
    }

    zap.L().Info("用户登录信息已保存到 Redis",
        zap.Uint("userID", userID),
        zap.String("username", username),
        zap.String("key", key))

    return nil
}
```

**变化**：
- 新增 `rdb *redis.Client` 字段
- 构造函数接收 Redis 客户端
- `Login` 方法调用 `saveLoginInfoToRedis`
- 新增 `saveLoginInfoToRedis` 方法实现 Redis 写入逻辑

---

## 🎯 功能特性

1. **Redis Key 格式**：`user:login:{userID}`
2. **数据格式**：JSON 字符串，包含 `user_id`、`username`、`access_token`、`login_time`
3. **过期时间**：7 天自动过期
4. **容错设计**：Redis 写入失败不阻塞登录流程，仅记录错误日志
5. **日志输出**：成功时输出 Info 日志，失败时输出 Error 日志

---

## 🚀 快速验证

### 1. 启动服务
```bash
cd /Users/tuliuxiang/Desktop/GITFilter/fullstack-app/server
make dev
```

### 2. 登录测试
```bash
curl -X POST http://localhost:6767/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"name":"admin","password":"admin123"}'
```

### 3. 查看 Redis
```bash
redis-cli
127.0.0.1:6379> KEYS user:login:*
127.0.0.1:6379> GET user:login:1
127.0.0.1:6379> TTL user:login:1
```

---

## 📝 相关文档

- 详细使用说明：`REDIS_LOGIN.md`
- 配置示例：`config/config.example.yaml`
- 项目文档：`CLAUDE.md`

---

## ✨ 编译状态

- ✅ 编译通过
- ✅ 无 lint 警告
- ✅ 测试通过（无测试文件）

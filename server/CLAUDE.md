# CLAUDE.md - Go 后端服务

## 项目概述

Go + React 全栈 Monorepo。本目录（`server/`）为 Go 后端服务，提供用户管理、角色权限、身份认证和文件上传等 RESTful API。

## 技术栈

- **语言**: Go 1.23
- **Web 框架**: Hertz (CloudWeGo，不是 Gin) - Handler 签名: `(ctx context.Context, c *app.RequestContext)`
- **ORM**: GORM v2 + MySQL 驱动
- **认证**: JWT (golang-jwt/v5, HS256, access + refresh 双令牌)
- **配置**: Viper (YAML 格式)
- **日志**: 标准库 `log/slog`，JSON 输出
- **缓存**: go-redis v9 (已连接但应用层暂未使用)
- **权限**: Casbin v2 (中间件已编写在 `middleware/casbin.go`，但未注册到路由)
- **校验**: Hertz 内置 `BindAndValidate`，使用 `vd:` 结构体标签
- **热重载**: Air (配置文件 `.air.toml`)

## 架构

清洁分层架构，在 `cmd/server/main.go` 中手动构造注入依赖：

```
Handler -> Service -> Repository -> Model
    |          |           |
    v          v           v
HTTP 绑定    业务逻辑     GORM 查询
+ 响应格式               + 数据访问
```

- `cmd/server/main.go` - 入口文件，依赖注入组装
- `internal/config/` - Viper 配置结构体 + 加载器
- `internal/database/` - MySQL (GORM) 和 Redis 连接初始化
- `internal/handler/` - HTTP 处理器 (auth, user, role, upload)
- `internal/service/` - 业务逻辑层 + 请求/响应 DTO
- `internal/repository/` - 数据访问层
- `internal/model/` - 领域实体 (GORM 模型)
- `internal/middleware/` - RequestID, Recovery, CORS, Logger, JWT, Casbin
- `internal/router/` - 路由注册
- `pkg/errcode/` - 类型化错误码
- `pkg/jwt/` - JWT 令牌管理器
- `pkg/response/` - 统一 JSON 响应工具
- `pkg/upload/` - 文件上传工具

## 常用命令

```bash
# 开发模式（在 server/ 目录下）
make dev          # 使用 Air 热重载运行
air               # 直接热重载

# 构建
make build        # CGO_ENABLED=0 go build -o bin/server cmd/server/main.go

# 测试
make test         # go test ./... -v -cover

# 代码检查
make lint         # golangci-lint run ./...

# 依赖管理
make tidy         # go mod tidy

# Swagger 文档生成
make swagger      # swag init -g cmd/server/main.go -o docs

# 从项目根目录执行
make dev          # 启动 MySQL+Redis 容器，然后同时运行前后端
make deploy       # Docker Compose 生产环境部署
```

## API 端点

所有接口在 `/api/v1` 下。统一响应格式: `{"code": int, "data": any, "message": string}`。

**公开接口**（无需认证）:
- `GET /health` - 健康检查
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 登录（返回 access + refresh 令牌）
- `POST /api/v1/auth/refresh-token` - 刷新令牌对

**受保护接口**（需要 JWT Bearer）:
- `GET|POST /api/v1/users` - 用户列表（分页、关键词搜索）/ 创建用户
- `GET /api/v1/users/profile` - 当前用户信息
- `GET|PUT|DELETE /api/v1/users/:id` - 用户 CRUD
- `GET|POST /api/v1/roles` - 角色列表（分页）/ 创建角色
- `GET /api/v1/roles/all` - 所有启用角色（不分页）
- `GET|PUT|DELETE /api/v1/roles/:id` - 角色 CRUD
- `POST /api/v1/upload` - 文件上传（multipart）

## 数据库

- MySQL 8.0，使用 GORM AutoMigrate 自动迁移（无手动 SQL 迁移文件）
- 数据表: `users`, `roles`, `user_roles`（多对多关联表）
- 软删除: 通过 `gorm.DeletedAt` 实现
- 迁移时禁用外键约束

## 配置说明

- 运行配置: `config/config.yaml`（已 gitignore）
- 模板文件: `config/config.example.yaml`
- 本地开发使用本机 MySQL: `root:123456789@tcp(127.0.0.1:3306)/fullstack_app`
- Docker 开发环境: 项目根目录 `docker-compose.yml`（MySQL + Redis）

## 编码规范

- **错误码**: 按领域划分: 10xxx=认证, 20xxx=用户, 30xxx=角色, 40xxx=上传。0 表示成功。
- **响应**: 统一使用 `pkg/response/` 中的 `response.OK()`, `response.Fail()`, `response.OKWithPage()`
- **请求 DTO**: 定义在 `internal/service/` 中，与 Service 方法放在一起
- **命名**: Go 文件 snake_case，包名小写单词，JSON 字段 snake_case，YAML 配置键 snake_case
- **中间件栈**: 全局: RequestID -> Recovery -> CORS -> Logger；JWT 仅用于受保护路由组
- **参数校验**: 使用 `vd:` 标签（Hertz go-tagexpr），不要写独立的校验逻辑
- **密码**: bcrypt 哈希，密码字段通过 `json:"-"` 标签对 JSON 隐藏
- **分页**: 请求参数 `page` + `page_size`，响应包装为 `PageResult{list, total, page, page_size}`

## 重要提醒

- 本项目使用 Hertz (CloudWeGo)，不是 Gin。不要使用 `*gin.Context`，应使用 `(ctx context.Context, c *app.RequestContext)`。
- Redis 在 `main.go` 中已连接但未传递给任何 Service（应用层暂未使用）。
- Casbin RBAC 中间件已编写但未注册到任何路由组。
- 目前没有测试文件。
- `config.yaml` 已被 gitignore，需从 `config.example.yaml` 复制并修改本地配置。

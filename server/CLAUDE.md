# CLAUDE.md - Go 后端服务

## 项目概述

Go + React 全栈 Monorepo。本目录（`server/`）为 Go 后端服务，提供用户管理、角色权限、身份认证、文件上传、金融持仓/交易数据管理、订阅数据以及 AI 对话代理等 RESTful API。

## 技术栈

- **语言**: Go 1.25
- **Web 框架**: Hertz (CloudWeGo，不是 Gin) — Handler 签名: `(ctx context.Context, c *app.RequestContext)`
- **ORM**: GORM v2 + MySQL 驱动
- **认证**: JWT (golang-jwt/v5, HS256, access + refresh 双令牌)
- **配置**: Viper (YAML 格式，支持 `APP_` 前缀环境变量覆盖)
- **日志**: zap (uber-go/zap)，彩色控制台输出 (fatih/color)
- **缓存**: go-redis v9 (已连接但应用层暂未使用)
- **权限**: Casbin v2 (中间件已编写在 `middleware/casbin.go`，但未注册到路由)
- **校验**: Hertz 内置 `BindAndValidate`，使用 `vd:` 结构体标签
- **热重载**: Air (配置文件 `.air.toml`)
- **其他**: jinzhu/copier (结构体拷贝), google/uuid

## 架构

清洁分层架构，在 `cmd/server/main.go` 中手动构造注入依赖：

```
Handler -> Service -> Repository -> Model
    |          |           |
    v          v           v
HTTP 绑定    业务逻辑     GORM 查询
+ 响应格式               + 数据访问
```

### 目录结构

- `cmd/server/main.go` — 入口文件，依赖注入组装
- `config/` — `config.yaml` (运行配置, gitignored) + `config.example.yaml` (模板)
- `internal/config/` — Viper 配置结构体 + 加载器
- `internal/database/` — MySQL (GORM) 和 Redis 连接初始化
- `internal/handler/` — HTTP 处理器 (auth, user, role, upload, energy, trade, jy_data, sse, ai, menu, enum, subscription, theme_content)
- `internal/service/` — 业务逻辑层 + 请求/响应 DTO
- `internal/repository/` — 数据访问层 (user, role, energy, jy_data, menu, subscription, theme_content)
- `internal/model/` — GORM 数据库模型 (user, sys_role, sys_menu, sys_user_role, sys_menu_role, image, xq_subscription, xq_theme_content, energy, summary, jy_data)
- `internal/middleware/` — RequestID, Recovery, CORS, Logger, JWT, Casbin
- `internal/router/` — 顶层 Setup + `v1/` 子包 (14 个路由文件) + `v2/` 子包 (测试接口)
  - `registrar.go` — 定义 `Registrar` 接口，各版本 Handlers 实现该接口自行注册路由
  - `public.go` — 无版本前缀的公共路由（`/health`、`/energy/asset` 油猴脚本直连）
- `pkg/errcode/` — 类型化错误码
- `pkg/jwt/` — JWT 令牌管理器
- `pkg/response/` — 统一 JSON 响应工具
- `pkg/upload/` — 文件上传工具
- `pkg/snowflake/` — 雪花 ID 生成器（用于菜单 ID）

## 常用命令

```bash
make dev          # 使用 Air 热重载运行
make build        # CGO_ENABLED=0 go build -o bin/server cmd/server/main.go
make test         # go test ./... -v -cover
make lint         # golangci-lint run ./...
make tidy         # go mod tidy
make swagger      # swag init -g cmd/server/main.go -o docs
```

## API 端点

所有接口在 `/api/v1` 下。统一响应格式: `{"code": int, "data": any, "message": string}`。

**公开接口**（无需认证）:
- `GET /health` — 健康检查
- `POST /api/v1/auth/register` — 用户注册
- `POST /api/v1/auth/login` — 登录（返回 access + refresh 令牌、user、role、menuRoles）
- `POST /api/v1/auth/refresh-token` — 刷新令牌对
- `POST /api/v1/energy/insert` — 批量插入持仓数据
- `GET /api/v1/jydata/latest` — 获取最新交易数据
- `POST /api/v1/jydata/list` — 按日期范围查询交易数据
- `POST /api/v1/sse/chat-messages` — AI 对话 (SSE 流式)
- `GET /api/v1/sse/chat-messages/:id` — 获取对话历史
- `GET /api/v1/ai/conversations` — 获取会话列表
- `GET /api/v1/enums/roles` — 角色枚举列表
- `POST|DELETE|GET /api/v1/subscriptions` — 订阅 CRUD（含 toggle/exists/detail/detail-table）
- `POST|PUT|DELETE|GET /api/v1/theme-contents` — 主题内容 CRUD（含 batch/search/timeline）

**受保护接口**（需要 JWT Bearer）:
- `POST /api/v1/auth/logout` — 登出
- `GET|POST /api/v1/users` — 用户列表（分页、关键词搜索）/ 创建用户
- `GET /api/v1/users/profile` — 当前用户信息
- `GET|PUT|DELETE /api/v1/users/:id` — 用户 CRUD
- `PUT /api/v1/users/:id/role` — 更新用户角色
- `POST|GET /api/v1/users/:id/roles` — 批量分配 / 查询用户角色
- `GET|POST /api/v1/roles` — 角色列表（分页）/ 创建角色
- `GET /api/v1/roles/all` — 所有启用角色（不分页）
- `GET|PUT|DELETE /api/v1/roles/:id` — 角色 CRUD
- `POST /api/v1/upload` — 文件上传（multipart）
- `POST /api/v1/trade/index` — 按日期范围查询交易汇总
- `POST /api/v1/trade/summary` — 单日交易汇总详情
- `GET|POST /api/v1/menus` — 菜单列表 / 批量添加
- `POST /api/v1/menus/role-binding` — 角色-菜单绑定
- `GET /api/v1/menus/role-binding/:roleId` — 按角色查菜单

## 数据库

- MySQL 8.0，使用 GORM AutoMigrate 自动迁移（无手动 SQL 迁移文件）
- 当前 AutoMigrate 模型: `user`, `sys_role`, `sys_menu`, `sys_user_role`, `sys_menu_role`, `xq_subscription`, `xq_theme_content`, `image`, `energy`, `summary`, `jy_data`
- `sys_role` 使用业务字段 `del_flag` 软删除，不使用 GORM `DeletedAt`。
- `user` 保留 GORM `DeletedAt`，密码字段必须保持 `json:"-"`。
- 迁移时禁用外键约束。
- 旧 Gin 项目 `/Users/tuliuxiang/Desktop/GITFilter/golang/gin/dal/modal` 只作为数据库模型参考，不直接复制业务层代码。

## 配置结构 (`internal/config/config.go`)

```yaml
server:   # port, mode (debug/release)
mysql:    # host, port, username, password, database, max_idle_conns, max_open_conns
redis:    # host, port, password, db
jwt:      # secret, access_expire (hours), refresh_expire (hours), issuer
upload:   # path, max_size (MB), allow_exts
cors:     # allow_origins
ai:       # base_url, token (外部 AI 服务，Dify 风格 API)
```

## 编码规范

- **注释**: 所有代码注释必须使用中文。
- **错误码**: 按领域划分: 10xxx=认证/用户, 20xxx=角色, 30xxx=菜单, 40xxx=订阅, 50xxx=主题内容。0 表示成功。
- **响应**: 统一使用 `pkg/response/` 中的 `response.OK()`, `response.Fail()`, `response.OKWithPage()`
- **请求 DTO**: 定义在 `internal/service/` 中，与 Service 方法放在一起
- **命名**: Go 文件 snake_case，包名小写单词，JSON 字段 camelCase（如 `pageSize`），YAML 配置键 snake_case
- **中间件栈**: 全局: RequestID → Recovery → CORS → Logger；JWT 仅用于受保护路由组
- **参数校验**: 使用 `vd:` 标签（Hertz go-tagexpr），不要写独立的校验逻辑
- **密码**: bcrypt 哈希，密码字段通过 `json:"-"` 标签对 JSON 隐藏。
- **登录响应**: `auth/login` 返回 `token`、`user`、`role`、`menuRoles`；`menuRoles[].link_url` 需要和前端 `web/src/router/section/nav-router.ts` 的 `path` 对齐。
- **分页**: 请求参数 `page` + `pageSize`，响应包装为 `PageResult{list, total, page, pageSize}`
- **路由注册**: 各版本 Handlers 实现 `Registrar` 接口，在 `Register` 方法中自行创建路由组并注册路由。`router.Setup` 中逐个调用各版本的 `Register`。新增版本时创建 `internal/router/vN/` 包并实现接口即可。
- **ID 生成**: 菜单等需要分布式 ID 的场景使用 `pkg/snowflake`

## 重要提醒

- 本项目使用 **Hertz (CloudWeGo)**，不是 Gin。不要使用 `*gin.Context`，应使用 `(ctx context.Context, c *app.RequestContext)`。
- 依赖注入是手动的，所有组装在 `main.go` 中完成，无 DI 框架。
- Repository 和 Service 是具体结构体，未抽取接口。
- Redis 在 `main.go` 中已连接但未传递给任何 Service（应用层暂未使用）。
- Casbin RBAC 中间件已编写但未注册到任何路由组。
- 目前没有测试文件。
- 修改后端后优先运行 `make -C server test`。
- `config.yaml` 已被 gitignore，需从 `config.example.yaml` 复制并修改本地配置。
- AutoMigrate 不会删除旧字段；涉及缩短 varchar 等字段长度时，先检查现有数据，避免 MySQL Data truncated 错误。
- AI 相关服务 (SseService, AiService) 代理到外部 Dify 风格 API，通过 `config.ai` 配置。

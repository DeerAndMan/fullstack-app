# internal/middleware

## 作用

Hertz 中间件集合。提供请求级横切能力（请求 ID、日志、CORS、Panic 恢复、鉴权等），由 `internal/router/router.go` 在 `Setup` 阶段挂载。

## 内容

- `requestid.go` — 注入唯一 `X-Request-Id`，写入 ctx 与响应头
- `recovery.go` — Panic 捕获，避免单个请求拖垮进程
- `cors.go` — 跨域，允许来源从 `config.cors.allow_origins` 读取
- `logger.go` — 访问日志（zap）
- `jwt.go` — Bearer token 解析与校验，把 `user_id` / `role_key` 写入 ctx
- `casbin.go` — RBAC 权限校验（**已实现但当前未挂载到任何路由组**）

## 约定

- 全局中间件挂载顺序固定：**RequestID → Recovery → CORS → Logger**，在 `router.Setup` 中按此顺序追加。
- JWT 仅用于受保护路由组（在 v1 路由文件里通过 `group.Use(jwt.Auth())` 单独挂载）。
- 中间件内部如需写响应，统一走 `pkg/response`，禁止直接 `c.JSON`。
- 中间件**不依赖**具体 Service / Repository；如需读用户信息从 ctx 取，不要从 DB 查。
- 启用 Casbin 时记得初始化 enforcer 并在 `router.Setup` 中挂载到对应路由组。

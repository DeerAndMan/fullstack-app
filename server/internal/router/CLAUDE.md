# internal/router

## 作用

路由注册中心。组装 Hertz 引擎、挂载中间件栈，并把各版本（v1 / v2）的 Handlers 注册到对应路径前缀下。

## 内容

- `router.go` — 总入口 `Setup(h *server.Hertz, deps ...)`：挂载全局中间件，调用各版本的 `Register`
- `registrar.go` — 定义 `Registrar` 接口，所有版本 Handlers 实现它来自行注册路由
- `public.go` — 无版本前缀的公共路由：`/health`、`/energy/asset`（油猴脚本直连）等
- `v1/` — `/api/v1` 业务路由（约 14 个文件，按领域拆分：auth / user / role / menu / upload / trade / jy_data / energy / sse / ai / enum / subscription / theme_content）
- `v2/` — `/api/v2` 路由：`ws.go`（WebSocket 升级）+ `routes.go`（含 test 端点）

## 约定

- 中间件挂载顺序（全局）：**RequestID → Recovery → CORS → Logger**。
- JWT 不放全局，只在受保护路由组上 `group.Use(jwt.Auth())`。
- 新增 API：在对应版本子包里实现 `Registrar.Register(rg *route.RouterGroup)`，由 `router.Setup` 统一调度，**不要**直接在 `router.go` 里散列地 `POST/GET`。
- 新增 API 版本：创建 `internal/router/vN/` 子包并实现 `Registrar` 即可。
- 路由路径与方法应当与前端 `web/src/api/api-control.ts` 保持同步；菜单路由还需与 `sys_menu.link_url` 一致。

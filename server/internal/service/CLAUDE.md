# internal/service

## 作用

业务逻辑层。聚合一个或多个 Repository，完成领域用例（如登录、菜单绑定、订阅 toggle、SSE 代理转发等）。同时承载请求/响应 DTO 定义。

## 内容

- 与 `internal/handler` 同名一一对应：`auth.go` / `user.go` / `role.go` / `menu.go` / `upload.go` / `trade.go` / `jy_data.go` / `energy.go` / `subscription.go` / `theme_content.go` / `sse.go` / `ai.go`
- `ws_hub.go` — WebSocket 连接池与广播枢纽，被 `handler/ws_v2.go` 持有；维护客户端集合、注册/注销、广播消息

## 约定

- Service 是具体 struct（未抽接口）；通过构造函数 `NewXxxService(repo, ...)` 由 `cmd/server/wire.go` 装配。
- **请求 / 响应 DTO** 定义在本目录（与方法放在同文件），handler 直接引用。
- 业务错误统一返回 `pkg/errcode` 的类型化错误码，按领域段位划分（10xxx / 20xxx ...）。
- 跨领域调用走 Service ↔ Service；**禁止**在 Service 中直接调 GORM，必须通过 Repository。
- 密码 bcrypt 在此层处理；模型字段 `Password` 必须保持 `json:"-"`。
- AI 类服务（`sse.go` / `ai.go`）代理到外部 Dify 风格 API，使用 `config.ai` 配置。
- `ws_hub.go` 是单例 + goroutine 模型，新增广播通道时注意 channel 关闭顺序，避免 panic。

## 与 handler / repository 的边界

| 维度          | handler                          | service                         | repository                |
| ------------- | -------------------------------- | ------------------------------- | ------------------------- |
| 关心的事      | HTTP 协议（路径、方法、Header、Body、状态码） | 业务规则与流程编排              | 数据库表结构与 GORM 查询  |
| 输入          | `*app.RequestContext`            | DTO 结构体                      | DTO / 主键 / 条件结构体   |
| 输出          | JSON 响应（走 `pkg/response`）   | DTO / `pkg/errcode` 业务错误码 | Model / 原始 `error`      |
| 能写 SQL/GORM | 不行                             | 不行                            | 唯一允许                  |
| 能读 HTTP 上下文 | 唯一允许                      | 不行                            | 不行                      |

判断「该不该写在 service」用以下规则：

1. 体现一条**业务规则**（"停用用户不能登录"、"订阅数量上限"、"密码必须 bcrypt"、"调用外部 AI 接口"），**是**就放 service。
2. 涉及 `c.Query / c.Bind / Cookie` 等 HTTP 上下文，**不要**放 service，挪回 handler。
3. 涉及 `gorm` 调用，**不要**放 service，挪去 repository。
4. **跨表 / 跨聚合根**的组合查询或事务编排，由 service 触发；事务函数体内的 GORM 操作仍在 repository。
5. 跨领域调用走 service ↔ service，**禁止**让一个 repository 调另一个 repository。

service 层最常见的错放：

| 反例                                          | 应该挪到   | 原因                                          |
| --------------------------------------------- | ---------- | --------------------------------------------- |
| service 里 `c.Query("id")` / `c.GetHeader(...)` | handler    | service 不依赖 HTTP 上下文                   |
| service 里手写 SQL 字符串或 `db.Raw(...)`      | repository | service 只调 repository 方法                  |
| service 里 `if err == gorm.ErrRecordNotFound` 后直接返回原始 error | 翻译成 `errcode.XxxNotFound` | service 的职责就是把 DB 错误翻译为业务错误码 |
| service A 直接调 repository B（B 与 A 无关）   | service A 调 service B  | 跨领域必须经 service，避免数据访问层耦合      |
| 把请求 DTO 定义在 handler 里                   | service    | DTO 与 service 方法同文件，handler 直接引用   |

> handler 是**翻译官**（HTTP ↔ Go），service 是**决策者**（业务规则），repository 是**搬运工**（DB ↔ Go）。

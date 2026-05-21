# internal/handler

## 作用

HTTP 处理层。只做参数绑定、校验和响应封装，**不写业务逻辑**。每个 handler 文件对应一个领域，调用同名 `internal/service` 完成实际工作。

## 目录结构

按 API 版本组织：

```
handler/
├── v1/              # /api/v1 版本的 handlers
│   ├── auth.go
│   ├── user.go
│   ├── role.go
│   ├── menu.go
│   ├── upload.go
│   ├── enum.go
│   ├── trade.go
│   ├── jy_data.go
│   ├── energy.go
│   ├── subscription.go
│   ├── theme_content.go
│   ├── sse.go      # SSE 流式 AI 对话
│   └── ai.go       # 会话历史
└── v2/              # /api/v2 版本的 handlers
    ├── test.go      # 测试端点
    └── ws.go        # WebSocket 接入点
```

## 内容

### v1 版本（业务主体）

- **业务**：`auth.go` / `user.go` / `role.go` / `menu.go` / `upload.go` / `enum.go`
- **数据**：`trade.go` / `jy_data.go` / `energy.go` / `subscription.go` / `theme_content.go`
- **实时 / AI**：`sse.go`（SSE 流式 AI 对话）/ `ai.go`（会话历史）

### v2 版本（实验性功能）

- `test.go` — 测试端点
- `ws.go` — WebSocket 接入点，配合 `service.WsHub` 实现服务端主动推送

## 约定

- **包名**：v1 目录下的文件 `package v1`，v2 目录下的文件 `package v2`
- **类型命名**：去掉版本后缀，如 `TestHandler`（不是 `TestHandlerV2`），通过包名区分版本
- Hertz 风格签名：`func (h *XxxHandler) Method(ctx context.Context, c *app.RequestContext)`，**不要**使用 `*gin.Context`。
- 参数绑定 + 校验用 `c.BindAndValidate(&req)`，配合 `vd:` 标签；不写独立的 if/return 校验。
- 响应统一走 `pkg/response`：`response.OK / Fail / OKWithPage`，禁止直接 `c.JSON`。
- 请求 / 响应 DTO 定义在 `internal/service` 中，handler 只引用、不重新定义。
- 文件命名 snake_case，与 service / repository 中的同领域文件保持一致。
- 新增领域时同步在 `internal/router/v1` 或 `v2` 下注册路由。
- 新增版本时创建 `internal/handler/vN/` 目录，并在 `internal/router/vN/` 中注册对应路由。

## 与 service / repository 的边界

| 维度          | handler                          | service                         | repository                |
| ------------- | -------------------------------- | ------------------------------- | ------------------------- |
| 关心的事      | HTTP 协议（路径、方法、Header、Body、状态码） | 业务规则与流程编排              | 数据库表结构与 GORM 查询  |
| 输入          | `*app.RequestContext`            | DTO 结构体                      | DTO / 主键 / 条件结构体   |
| 输出          | JSON 响应（走 `pkg/response`）   | DTO / `pkg/errcode` 业务错误码 | Model / 原始 `error`      |
| 能写 SQL/GORM | 不行                             | 不行                            | 唯一允许                  |
| 能读 HTTP 上下文 | 唯一允许                      | 不行                            | 不行                      |

判断「该不该写在 handler」用以下规则：

1. 涉及 `c.JSON / c.Bind / Header / Cookie / 状态码`，**是**就放 handler。
2. 涉及 `gorm` 调用 / 查询条件构造，**不要**放 handler，挪去 repository。
3. 体现业务规则（如"停用的用户不能登录"），**不要**放 handler，挪去 service。
4. 参数格式校验（必填、长度、枚举），用 service DTO 上的 `vd:` 标签 + `BindAndValidate` 触发，handler 不写 if/return 校验。
5. 权限判断走 `middleware/jwt`、`casbin`，**不在** handler 里。

handler 层最常见的错放：

| 反例                                | 应该挪到   | 原因                                |
| ----------------------------------- | ---------- | ----------------------------------- |
| handler 里 `db.Where(...).Find(...)` | repository | handler 不碰 ORM                    |
| handler 里 `if user.Status == 0 { 报错 }` | service    | 业务规则归 service                  |
| handler 里 `c.JSON(200, gin.H{...})` | 改用 `response.OK` | 统一响应格式由 `pkg/response` 保证 |
| handler 里手写 `if req.Name == ""` 校验 | DTO 加 `vd:` 标签 | 校验靠 tag，不靠手写代码           |
| handler 里直接构造业务错误码消息    | service    | 错误码翻译由 service 完成           |

> handler 是**翻译官**（HTTP ↔ Go），service 是**决策者**（业务规则），repository 是**搬运工**（DB ↔ Go）。

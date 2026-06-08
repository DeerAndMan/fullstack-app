# internal/repository

## 作用

数据访问层。封装 GORM 查询、事务、条件构造，是 Service 唯一允许访问数据库的入口。

## 内容

- `user.go` / `role.go` / `menu.go` — RBAC 三件套
- `energy.go` / `jy_data.go` — 金融持仓 / 交易数据
- `subscription.go` / `theme_content.go` — 订阅与主题内容

## 约定

- 每个 Repository 是具体 struct，构造函数 `NewXxxRepository(db *gorm.DB)`；未抽接口。
- 方法签名通常返回 `(result, error)`，错误直接返回 `gorm.ErrRecordNotFound` 等原始错误，**不要**在此层包装为业务错误码（错误码由 Service 翻译）。
- 软删除：`sys_role` 用业务字段 `del_flag`，**不要**给它加 GORM `DeletedAt`；`user` 保留 GORM `DeletedAt`。
- 复杂查询写在 Repository，Service 只调方法、不拼 SQL。
- 事务：跨表写操作用 `db.Transaction(func(tx *gorm.DB) error { ... })`，事务边界由 Service 触发、在 Repository 内执行。
- 分页通用模式：`Offset((page-1)*pageSize).Limit(pageSize)` + 同条件的 `Count`，结果由 Service 包装成 `PageResult`。

## 与 handler / service 的边界

| 维度          | handler                          | service                         | repository                |
| ------------- | -------------------------------- | ------------------------------- | ------------------------- |
| 关心的事      | HTTP 协议（路径、方法、Header、Body、状态码） | 业务规则与流程编排              | 数据库表结构与 GORM 查询  |
| 输入          | `*app.RequestContext`            | DTO 结构体                      | DTO / 主键 / 条件结构体   |
| 输出          | JSON 响应（走 `pkg/response`）   | DTO / `pkg/errcode` 业务错误码 | Model / 原始 `error`      |
| 能写 SQL/GORM | 不行                             | 不行                            | 唯一允许                  |
| 能读 HTTP 上下文 | 唯一允许                      | 不行                            | 不行                      |

判断「该不该写在 repository」用以下规则：

1. 涉及 `db.Find / .Where / .Create / .Updates / .Transaction` 等 GORM 调用，**是**就放 repository。
2. 涉及 HTTP 上下文或 `c.*`，**不在** repository 里出现。
3. 涉及业务规则（如"停用用户不能登录"、"订阅数上限"），**不要**写在 repository 里，挪去 service。
4. 错误码翻译（`gorm.ErrRecordNotFound → errcode.UserNotFound`）由 service 完成；repository 直接返回原始 error。
5. **一个 repository 只服务一个聚合根**（一张主表 + 强关联子表），跨表组合走 service。

repository 层最常见的错放：

| 反例                                                    | 应该挪到 | 原因                                            |
| ------------------------------------------------------- | -------- | ----------------------------------------------- |
| repository 里 `return errcode.UserNotFound`            | service  | repository 不感知业务错误码                     |
| repository 里 `if user.Status == 0 { 报错 }`           | service  | 业务规则不属于数据层                            |
| repository A 内部调 repository B                        | service  | 跨聚合根的组合由 service 编排，避免数据层耦合   |
| repository 里 `db.Begin() / .Commit()` 跨表事务         | service 触发，repository 用传入的 `tx` 执行 | 事务边界由业务定义                |
| repository 里读取 `c.GetString("user_id")`              | handler/service 传参 | repository 不应访问请求上下文           |
| 在 repository 里包装 `fmt.Errorf("用户不存在: %w", err)` | service 翻译 | repository 透传原始 error，让 service 决定如何翻译 |

> handler 是**翻译官**（HTTP ↔ Go），service 是**决策者**（业务规则），repository 是**搬运工**（DB ↔ Go）。

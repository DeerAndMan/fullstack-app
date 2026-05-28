# internal/database

## 作用

数据库与缓存的连接初始化、生命周期管理。被 `cmd/server` 在启动阶段调用，返回的连接对象注入到 Repository / Service。

## 内容

- `mysql.go` — GORM v2 + MySQL 驱动初始化；构造 `*gorm.DB`；负责日志配置、连接池参数（`max_idle_conns`、`max_open_conns`）、`AutoMigrate` 调用
- `redis.go` — go-redis v9 客户端初始化；构造 `*redis.Client`

## 约定

- 只做「建立连接」和「健康检查」，**不写**业务查询；业务查询放在 `internal/repository`。
- AutoMigrate 在此处统一注册模型，**禁止**在其他层调用。
- 迁移时禁用外键约束（`DisableForeignKeyConstraintWhenMigrating: true`）。
- 缩短 varchar 字段长度前先检查现网数据，避免 MySQL `Data truncated` 错误。
- Redis 连接虽然已初始化，但应用层尚未广泛使用，新增使用时记得在 Service 层注入。

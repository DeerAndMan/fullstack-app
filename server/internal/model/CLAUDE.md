# internal/model

## 作用

GORM 实体模型层。**只描述表结构**，不包含业务方法和查询逻辑（查询走 `internal/repository`）。

## 内容

- `base.go` — 公共字段（`ID` / `CreatedAt` / `UpdatedAt` 等基础嵌入结构）
- 业务表：`user.go` / `role.go` / `menu.go` / `sys_user_role.go` / `sys_menu_role.go` / `image.go`
- 数据表：`energy.go` / `summary.go` / `jy_data.go`
- 订阅相关：`xq_subscription.go` / `xq_theme_content.go`

## 约定

- 表名与字段名以**当前 MySQL 表结构**为准；参考旧 Gin 项目 `/Users/tuliuxiang/Desktop/GITFilter/golang/gin/dal/modal` 时只取模型，**不要**搬业务方法。
- 字段命名：DB 列 snake_case，Go 字段 PascalCase，JSON tag camelCase。
- **敏感字段**（如 `User.Password`）必须 `json:"-"`，避免接口泄漏 bcrypt hash。
- `sys_role` 软删除使用业务字段 `del_flag`；**不要**改成 GORM `DeletedAt`。
- 新增模型时记得在 `internal/database/mysql.go` 的 `AutoMigrate` 列表里注册。
- AutoMigrate 只做兼容性迁移、不会删除旧字段；缩短 varchar 等字段长度前先核对线上数据长度。

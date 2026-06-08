# internal/config

## 作用

应用配置的结构体定义与加载逻辑，负责把 `server/config/config.yaml` 解析为 Go 结构体并交给上层使用。

## 内容

- `config.go` — Viper 加载器 + 配置结构体（`Server` / `MySQL` / `Redis` / `JWT` / `Upload` / `CORS` / `AI` 等子结构）

## 约定

- 配置文件路径以 `server/config/config.yaml` 为准，模板见 `config.example.yaml`。
- 支持 `APP_` 前缀的环境变量覆盖（Viper `AutomaticEnv`）。
- YAML 键使用 snake_case，结构体字段使用 PascalCase 并加 `mapstructure` tag。
- **不要**在此目录写业务逻辑；只做配置加载与结构定义。
- 新增配置项时同步更新 `config.example.yaml`，并在根 `CLAUDE.md` 的「配置结构」段补充说明。

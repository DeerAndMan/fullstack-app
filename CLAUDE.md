# CLAUDE.md - Fullstack App

## 项目概述

这是一个 Go + React 全栈 Monorepo：

- `server/`：Go 后端服务，提供认证、用户、角色、上传、金融数据、订阅数据和 AI/SSE 相关接口。
- `web/`：React 前端应用，提供登录、导航、用户/角色管理、交易数据、订阅、SSR Demo 等页面。
- `deploy/`：Docker、Nginx 和生产部署编排。

优先阅读子目录内的 `CLAUDE.md`：

- 后端任务看 `server/CLAUDE.md`
- 前端任务看 `web/CLAUDE.md`

## 常用命令

```bash
make dev            # 启动 MySQL/Redis，并并行启动前后端开发服务
make build          # 构建前后端
make test           # 运行前后端测试
make lint           # 运行前后端 lint
make deploy         # 使用 deploy/docker-compose.prod.yml 生产部署
make clean          # 清理构建产物
```

分开执行：

```bash
make -C server test         # 后端测试
npm --prefix web run build  # 前端类型检查 + 生产构建
npm --prefix web run dev    # 前端 Vite 开发服务，端口 6565
```

## 技术栈

### 后端

- Go 1.25
- Hertz，不是 Gin
- GORM v2 + MySQL
- Viper 配置
- JWT access/refresh 双令牌
- Redis 已初始化但业务层暂未广泛使用
- Casbin 中间件存在但未注册到路由

### 前端

- React 19 + TypeScript 5.8
- Vite 8 + SWC
- Ant Design 5 + Tailwind CSS 4
- Zustand 持久化认证状态
- TanStack Query
- React Router 7
- Axios 请求封装

## 项目结构

```text
fullstack-app/
├── server/                  # Go 后端
│   ├── cmd/server/main.go   # 后端入口和依赖注入
│   ├── config/              # config.example.yaml + 本地 config.yaml
│   ├── internal/            # handler/service/repository/model/router/middleware
│   └── pkg/                 # errcode/jwt/response/upload/snowflake
├── web/                     # React 前端
│   ├── src/api/             # Axios 实例、接口路径、Zod 校验请求、领域 API (user/trade/menu/subscribe/enum)
│   ├── src/components/      # 通用组件 (chart/form/toast)
│   ├── src/hooks/           # 自定义 hooks (useBoolean)
│   ├── src/layouts/         # Layout 守卫 + Nav 导航
│   ├── src/pages/           # 页面 (home/login/user/role/trade/subscribe/ws/ssr-demo)
│   ├── src/router/          # 静态路由、懒加载、导航菜单、SPA/SSR 路由配置
│   ├── src/sections/        # 页面子模块 (subscribe/user)
│   ├── src/stores/          # Zustand store (auth/enum/global)
│   ├── src/theme/           # Ant Design 主题和明暗切换
│   ├── src/types/           # TS 类型、Zod schema、业务枚举
│   └── src/utils/           # cookie、图片、RSA 加密、树转换、WebSocket
├── deploy/                  # 部署配置
├── docker-compose.yml       # 本地 MySQL/Redis
└── Makefile                 # 顶层命令
```

## 前后端联动重点

- API 基础路径由 `web/.env.*` 的 `VITE_WEB_BASE_URL` 控制。
- 本地前端默认端口是 `6565`，后端实际端口以 `server/config/config.yaml` 为准（默认 `6767`）。
- 前端接口路径集中在 `web/src/api/api-control.ts`，所有路径已统一使用 `/api/v1` 前缀。
- 后端接口统一挂在 `/api/v1`。
- 统一响应格式是 `{"code": number, "data": any, "message": string}`，`code === 0` 表示成功。
- 分页响应格式是 `{"code": 0, "data": {"list": [], "total": N, "page": N, "pageSize": N}, "message": "success"}`。
- 前端分页类型 `PageData<T>` 定义在 `web/src/api/request.ts`，字段使用 camelCase（`pageSize` 而非 `page_size`）。
- 登录接口返回 `token`、`user`、`role`、`menuRoles`。
- 前端导航 `web/src/layouts/Nav.tsx` 使用 `role.role_key` 和 `menuRoles[].link_url` 过滤菜单。
- `sys_menu.link_url` 必须和 `web/src/router/section/nav-router.ts` 中的 `path` 完全一致，否则非超管登录后不会显示对应菜单。
- 前端请求层有两套：普通 Axios（`request.get/post`）和 Zod 校验版（`RequestSchema/RequestGet/RequestPost`），新接口优先使用 Zod 校验版。

## 开发约定

- 后端保持 Handler -> Service -> Repository -> Model 分层。
- 前端接口类型和后端响应字段要同步更新。
- 不要把旧 Gin 项目的代码直接迁移进当前后端；旧项目只能作为数据库模型和业务逻辑参考。
- 后端密码字段必须保持 `json:"-"`，避免接口泄露 bcrypt hash。
- 数据库模型以当前 MySQL 表结构为准，参考旧项目 `/Users/tuliuxiang/Desktop/GITFilter/golang/gin/dal/modal` 时要保留当前服务的安全差异。
- 修改前端 UI 后，优先运行 `npm --prefix web run build`，必要时启动开发服务手动验证。
- 修改后端后，优先运行 `make -C server test`。

## 注意事项

- 不要在当前项目中使用 Gin handler 写法；后端是 Hertz。
- 不要把 `.env`、`config.yaml`、密钥、数据库密码等敏感文件提交。
- `server/config/config.yaml` 被 gitignore，本地运行前需要从 `config.example.yaml` 复制。
- AutoMigrate 只做兼容性迁移，不会删除旧字段；涉及缩短字段长度时要先检查线上数据长度。

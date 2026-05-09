# CLAUDE.md - React 前端应用

## 项目概述

本目录（`web/`）是 `fullstack-app` 的 React 前端应用，负责登录、布局导航、用户/角色管理、交易数据展示、订阅功能、WebSocket 页面和 SSR Demo。

前端通过 Axios 调用 Go 后端接口，认证状态由 Zustand 持久化到 localStorage，同时写入认证 Cookie。

## 技术栈

- React 19
- TypeScript 5.8，严格模式开启
- Vite 8 + `@vitejs/plugin-react-swc`
- Ant Design 5 + @ant-design/charts 2
- Tailwind CSS 4
- Zustand 5
- TanStack Query 5
- React Router 7
- Axios + axios-retry
- Zod 4（接口响应校验）
- ts-pattern 5（模式匹配）
- node-forge（RSA 密码加密）
- decimal.js（精确数值计算）
- dayjs（日期处理）
- Express SSR 入口保留
- 包管理器: **pnpm**（锁文件为 `pnpm-lock.yaml`）

## 常用命令

```bash
pnpm --prefix web dev        # 启动 Vite 开发服务，默认端口 6565
pnpm --prefix web build      # TypeScript 构建检查 + Vite 生产构建
pnpm --prefix web lint       # ESLint
pnpm --prefix web test       # Vitest
pnpm --prefix web dev:ssr    # SSR 开发模式
pnpm --prefix web build:ssr  # SSR 构建
```

如果已经在 `web/` 目录内：

```bash
pnpm dev
pnpm build
pnpm lint
pnpm test
```

## 目录结构

- `src/main.tsx` — CSR 入口
- `src/App.tsx` — 应用根组件
- `src/antd-context.tsx` — Ant Design 全局上下文
- `src/api/` — Axios 实例、接口路径、Zod 校验请求和领域 API
  - `api-control.ts` — 所有接口路径常量（已统一 `/api/v1` 前缀）
  - `request.ts` — Axios 实例 + 拦截器 + `ResponseData<T>` / `PageData<T>` 类型
  - `request-schema.ts` — Zod 校验请求封装（`RequestSchema`/`RequestGet`/`RequestPost`/`RequestPut`/`RequestDelete`）
  - `user/` — 用户/认证 API + TanStack Query hooks
  - `trade/` — 交易数据 API + Query hooks
  - `menu/` — 菜单管理 API + Query/Mutation hooks
  - `subscribe/` — 订阅管理 API + Query hooks（home + detail）
  - `enum/` — 枚举 API + Query hooks
- `src/components/` — 通用组件 (chart: LineChart/DualAxesChart, form: FormWrap/FormItem, toast: Toastify, query: QueryProvider)
- `src/hooks/` — 自定义 hooks (useBoolean)
- `src/layouts/` — 页面布局、登录守卫和顶部导航
- `src/pages/` — 页面组件 (home/login/user/role/trade/subscribe/ws/ssr-demo)
  - `role/` — 角色列表 (list.tsx) + 角色菜单绑定 (Menu.tsx)
  - `subscribe/` — 首页 (home) + 详情 (detail) + 列表 (list)
  - `user/` — 用户管理 + 菜单设置 (MenuSetting)
- `src/router/` — 路由注册、懒加载、导航菜单、SPA/SSR 路由配置
- `src/sections/` — 页面子模块 (subscribe: home table + detail, user: AddEditUserModal + AddEditUserRoleModal, tree: utils)
- `src/stores/` — Zustand store (auth: 认证持久化, enum: 角色枚举, global: messageApi)
- `src/theme/` — Ant Design 主题和明暗切换
- `src/types/` — TypeScript 类型、Zod schema、业务枚举
  - `user.ts` — Account/Role/MenuRole 类型
  - `menu-router.ts` — MenuItemType/TreeMenuItemType/RoleRoutingType
  - `enum.ts` — RoleKey 枚举
  - `schema.ts` — Zod schema (TradeItem/EnergyItem 等)
  - `schema-fixed.ts` — 修正版 Zod schema
  - `api.d.ts` — ApiResponse/PageResult/PageParams 通用类型
  - `constants.ts` — Any/CallbackFunction 工具类型
  - `global.d.ts` — 全局类型声明
  - `xq/subscribe/home.ts` — 订阅相关 Zod schema
- `src/utils/` — cookie、debugger、图片 base64、RSA 加密、树转换、WebSocket 封装
- `server.ts`、`server/`、`src/ssr-entry.tsx` — SSR 相关入口

## API 与请求约定

- API baseURL 来自 `import.meta.env.VITE_WEB_BASE_URL`。
- 本地开发配置在 `.env.development`，当前指向 `http://127.0.0.1:6767`。
- 生产配置在 `.env.production`。
- 接口路径集中维护在 `src/api/api-control.ts`，所有路径已统一使用 `/api/v1` 前缀。
- Axios 实例在 `src/api/request.ts`。
- 请求默认带 `Authorization: Bearer <token>`；传 `noToken` 时不带 token。
- 后端统一返回：`{ code, data, message }`。
- `code === 0` 表示成功；非 0 会被响应拦截器 reject。
- `401/403` 或业务 token 错误码会清空认证状态。
- 分页响应使用 `PageData<T>` 类型：`{ list: T[], total, page, pageSize }`，字段统一 camelCase。
- 新接口优先使用 Zod 校验版请求函数（`RequestSchema`/`RequestGet`/`RequestPost`），定义在 `src/api/request-schema.ts`。

## 认证和导航

关键文件：

- `src/pages/login/index.tsx` — 登录页
- `src/stores/auth.ts` — token/user/role/menuRoles 持久化
- `src/layouts/Layout.tsx` — token 守卫
- `src/layouts/Nav.tsx` — 顶部导航渲染
- `src/router/section/nav-router.ts` — 静态导航配置

登录成功后必须保存：

- `token.access_token`
- `user`
- `role`
- `menuRoles`

导航显示规则：

- `role.role_key === RoleKey.SUPER_ADMIN` 时显示全部静态导航。
- 非超管使用 `menuRoles[].link_url` 匹配 `nav-router.ts` 中的 `path`。
- 如果数据库 `sys_menu.link_url` 和前端 `path` 不一致，导航不会显示对应菜单。

## 类型约定

- 用户和角色类型在 `src/types/user.ts`。
- 菜单权限类型在 `src/types/menu-router.ts`。
- 角色枚举在 `src/types/enum.ts`。
- 通用 API 类型（`ApiResponse`/`PageResult`/`PageParams`）在 `src/types/api.d.ts`。
- 交易/持仓 Zod schema 在 `src/types/schema.ts`。
- 订阅相关 Zod schema 在 `src/types/xq/subscribe/home.ts`。
- 修改后端响应字段时，要同步更新 `src/api/*` 的返回类型和 `src/types/*`。
- 当前 TypeScript 开启 `strict`、`noUnusedLocals`、`noUnusedParameters`、`noImplicitAny`，不要留下未使用变量或隐式 any。
- 路径别名：`@/*` → `./src/*`，另有 `@api/*`、`@store/*`、`@components/*`、`@pages/*`、`@router/*`、`@types/*`、`@layouts/*`、`@utils/*`。

## 路由约定

- 页面路由静态注册，不依赖后端动态路由。
- 导航菜单和路由页面是两套配置：路由负责能否访问页面，导航负责是否显示入口。
- 增加新页面时通常需要同时更新：
  - `src/router/section/router-list.tsx` 或相关路由文件
  - `src/router/section/nav-router.ts`
  - 后端/数据库中的菜单权限数据（如需要按角色显示）

## UI 开发约定

- 优先复用 Ant Design 组件。
- Tailwind class 主要用于布局、间距和颜色。
- 主题相关逻辑放在 `src/theme/`。
- 修改 UI 后应启动开发服务在浏览器验证关键路径；无法手动验证时要明确说明。
- 头像由后端返回 base64，前端通过 `src/utils/img.ts` 转为可显示图片。

## 主题系统

- 主题切换基于 React Context (`src/theme/antd-context.tsx`)，提供 `ThemeProvider` 和 `useTheme` hook。
- 支持 `light` / `dark` 两种模式，持久化到 `localStorage("theme")`。
- 切换时同步更新 `document.body.className` 和 CSS 变量（`--app-color-primary`、`--app-color-bg-base` 等）。
- Ant Design 的 `ConfigProvider` 根据主题切换 `lightTheme` / `darkTheme` 配置（定义在 `src/theme/antd-theme.ts`）。
- 全局 locale 固定为 `zhCN`，dayjs 也加载了 `zh-cn`。
- 主题切换组件: `src/theme/theme-switch.tsx`（Switch 组件，🌙/☀️ 图标）。

## 路由路径常量

路由路径集中定义在 `src/router/section/router-path.ts`：

```typescript
ROUTER_PATH = {
  login: "/login",
  home: "/",
  data: "/data",
  ws: "/ws",
  subscribe: { home, list, detail: "/subscribe/detail/:id/:userId" },
  role: { list: "/role/list", menu: "/role/menu" },
  user: { root: "/user", operation: "/user/operation" },
}
```

页面组件中引用路径时应使用此常量，避免硬编码字符串。

## Sections 模块详情

`src/sections/` 存放页面级别的子模块组件，按业务域组织：

- `subscribe/home/table.tsx` — 订阅首页表格组件
- `subscribe/home/operable-dialog.tsx` — 订阅操作弹窗
- `subscribe/detail/index.tsx` — 订阅详情子模块
- `user/AddEditUserModal.tsx` — 用户新增/编辑弹窗
- `user/AddEditUserRoleModal.tsx` — 用户角色分配弹窗
- `tree/utils.ts` — 树形数据处理工具函数

Sections 和 Pages 的区别：Pages 是路由级组件，Sections 是 Pages 内部的可复用子模块。

## SSR 架构

项目保留了 Express 5 + Vite SSR 的完整入口：

- `server.ts` — Express SSR 服务入口（开发模式使用 Vite middleware，生产模式读取构建产物）
- `src/ssr-entry.tsx` — SSR 渲染入口（`renderToString`）
- `src/ssr-context.tsx` — SSR 数据注入上下文
- `src/router/section/ssr-routes.ts` — SSR 专用路由配置
- `src/pages/ssr-demo/` — SSR 演示页面（基础、性能测试、数据获取）

SSR 构建命令：`pnpm build:ssr`（分别构建 client 和 server 产物到 `dist/`）。

当前主要使用 CSR 模式，SSR 仅作为技术演示保留。

## Axios 拦截器细节

请求拦截器支持的自定义配置项（`CustomRequestConfig`）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `noToken` | `boolean` | 不附加 Authorization header |
| `errorMsg` | `boolean \| string` | 错误时是否/如何显示 message（默认 true） |
| `successMsg` | `boolean \| string` | 成功时是否/如何显示 message |
| `saltLength` | `number` | 设置 `SaltLength` 请求头（RSA 加密相关） |
| `needRetry` | `boolean` | 500 错误时是否重试（最多 3 次，间隔递增） |
| `schema` | `z.ZodSchema` | Zod 响应校验 schema |
| `schemaType` | `"base" \| "table"` | schema 包装类型（table 会自动包装分页结构） |

响应拦截器行为：
- 检测 `x-new-token` header 自动刷新 token（静默续期）。
- `code === 0` 成功；非 0 reject 并可选显示错误消息。
- `tokenErrList` (401, 10002, 10003) 触发清空认证状态。
- HTTP 状态码 401/403 同样触发清空认证。
- Blob 响应直接透传，不做 code 判断。

## 导航菜单结构

`nav-router.ts` 定义的完整导航树：

```
首页 (/)
数据 (/data)
ws (/ws)
SSR Demo
  ├── 基础演示 (/ssr)
  ├── 性能测试 (/ssr/performance)
  └── 数据获取 (/ssr/data-fetch)
订阅
  ├── 订阅首页 (/subscribe)
  └── 订阅列表 (/subscribe/list)
管理
  └── 菜单
  │   └── 角色菜单管理 (/role/menu)
  └── 权限列表 (/role/list)
```

注意：`/subscribe/detail/:id/:userId` 和 `/user`、`/user/operation` 不在导航中显示，通过页面内链接跳转。

## 注意事项

- 不要把后端字段名改成前端自造字段；优先同步 TypeScript 类型适配真实接口。
- 不要绕过 `src/api/request.ts` 直接创建新的 Axios 实例，除非确有独立 baseURL/拦截器需求。
- 不要在登录页只保存 token/user；导航权限依赖 role/menuRoles。
- 前后端分页字段统一使用 camelCase（`pageSize`），不要用 `page_size`。
- `src/utils/encrypted.ts` 中的 RSA 公钥用于密码加密传输，修改时需同步后端解密逻辑。
- 主题切换状态独立于认证状态，不会因登出而重置。
- `src/api/auth.ts` 是旧的认证 API 文件，当前认证逻辑已迁移到 `src/api/user/` 模块。
- Zod 校验版请求在 schema 验证失败时会 `console.warn` 并 throw，开发时注意控制台警告。

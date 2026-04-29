# CLAUDE.md - React 前端应用

## 项目概述

本目录（`web/`）是 `fullstack-app` 的 React 前端应用，负责登录、布局导航、用户/角色管理、交易数据展示、订阅功能、WebSocket 页面和 SSR Demo。

前端通过 Axios 调用 Go 后端接口，认证状态由 Zustand 持久化到 localStorage，同时写入认证 Cookie。

## 技术栈

- React 19
- TypeScript 5.8，严格模式开启
- Vite 8 + `@vitejs/plugin-react-swc`
- Ant Design 5
- Tailwind CSS 4
- Zustand 5
- TanStack Query 5
- React Router 7
- Axios + axios-retry
- Zod
- Express SSR 入口保留

## 常用命令

```bash
npm --prefix web run dev        # 启动 Vite 开发服务，默认端口 6565
npm --prefix web run build      # TypeScript 构建检查 + Vite 生产构建
npm --prefix web run lint       # ESLint
npm --prefix web run test       # Vitest
npm --prefix web run dev:ssr    # SSR 开发模式
npm --prefix web run build:ssr  # SSR 构建
```

如果已经在 `web/` 目录内：

```bash
npm run dev
npm run build
npm run lint
npm run test
```

## 目录结构

- `src/main.tsx` — CSR 入口
- `src/App.tsx` — 应用根组件
- `src/antd-context.tsx` — Ant Design 全局上下文
- `src/api/` — Axios 实例、接口路径、请求类型和领域 API
- `src/components/` — 通用组件
- `src/hooks/` — 自定义 hooks 和 React Query hooks
- `src/layouts/` — 页面布局、登录守卫和顶部导航
- `src/pages/` — 页面组件
- `src/router/` — 路由注册、懒加载和导航菜单配置
- `src/sections/` — 页面子模块
- `src/stores/` — Zustand store
- `src/theme/` — Ant Design 主题和明暗切换
- `src/types/` — TypeScript 类型和业务枚举
- `src/utils/` — cookie、图片、加密、树转换等工具
- `server.ts`、`server/`、`src/ssr-entry.tsx` — SSR 相关入口

## API 与请求约定

- API baseURL 来自 `import.meta.env.VITE_WEB_BASE_URL`。
- 本地开发配置在 `.env.development`，当前指向 `http://127.0.0.1:6767`。
- 生产配置在 `.env.production`。
- 接口路径集中维护在 `src/api/api-control.ts`。
- Axios 实例在 `src/api/request.ts`。
- 请求默认带 `Authorization: Bearer <token>`；传 `noToken` 时不带 token。
- 后端统一返回：`{ code, data, msg/message }`。
- `code === 0` 表示成功；非 0 会被响应拦截器 reject。
- `401/403` 或业务 token 错误码会清空认证状态。

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
- 修改后端响应字段时，要同步更新 `src/api/*` 的返回类型和 `src/types/*`。
- 当前 TypeScript 开启 `strict`、`noUnusedLocals`、`noUnusedParameters`、`noImplicitAny`，不要留下未使用变量或隐式 any。

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

## 注意事项

- 不要把后端字段名改成前端自造字段；优先同步 TypeScript 类型适配真实接口。
- 不要绕过 `src/api/request.ts` 直接创建新的 Axios 实例，除非确有独立 baseURL/拦截器需求。
- 不要在登录页只保存 token/user；导航权限依赖 role/menuRoles。
- `src/api/api-control.ts` 里还有部分旧接口路径没有 `/api/v1` 前缀，接入当前 Go 后端时要特别核对。

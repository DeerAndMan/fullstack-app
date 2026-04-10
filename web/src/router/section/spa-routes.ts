/**
 * SPA 路由路径表（纯数据，不含 JSX）
 *
 * 用于 server.ts 启动时打印路由列表，与 ssr-routes.ts 配合展示全量路由。
 * 注意：此文件需与 router-list.tsx 中的路由保持同步。
 */

export interface SPARouteItem {
  /** 路由路径 */
  path: string;
  /** 页面标题 */
  title: string;
}

export const spaRoutes: SPARouteItem[] = [
  { path: "/", title: "首页" },
  { path: "/login", title: "登录" },
  { path: "/user", title: "用户详情" },
  { path: "/user/operation", title: "操作日志" },
  { path: "/data", title: "数据" },
  { path: "/ws", title: "WebSocket" },
  { path: "/subscribe", title: "订阅首页" },
  { path: "/subscribe/list", title: "订阅列表" },
  { path: "/subscribe/detail/:id/:userId", title: "订阅详情" },
  { path: "/role/menu", title: "角色菜单管理" },
  { path: "/role/menu/:id/add", title: "添加菜单" },
  { path: "/role/list", title: "权限列表" },
];

/** 所有 SPA 路径集合 */
export const spaPaths: string[] = spaRoutes.map((r) => r.path);

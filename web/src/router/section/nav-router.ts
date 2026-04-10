import type { TreeDataNode } from "antd";

export interface NavRouter {
  id: string;
  name: string;
  path: string;
  show?: boolean;
  children?: NavRouter[];
}

export type TreeDataType = TreeDataNode & {
  link_url: string;
};

/** 导航栏 */
export const navRouter: NavRouter[] = [
  { id: "home", name: "首页", path: "/" },
  { id: "data", name: "数据", path: "/data" },
  { id: "ws", name: "ws", path: "/ws" },
  {
    id: "ssr",
    name: "SSR Demo",
    path: "",
    children: [
      { id: "ssr-basic", name: "基础演示", path: "/ssr" },
      { id: "ssr-performance", name: "性能测试", path: "/ssr/performance" },
      { id: "ssr-data-fetch", name: "数据获取", path: "/ssr/data-fetch" },
    ],
  },
  {
    id: "subscribe",
    name: "订阅",
    path: "",
    children: [
      { id: "subscribe-home", name: "订阅首页", path: "/subscribe" },
      { id: "subscribe-list", name: "订阅列表", path: "/subscribe/list" },
    ],
  },
  {
    id: "auth",
    name: "管理",
    path: "",
    children: [
      {
        id: "auth-menu",
        name: "菜单",
        path: "",
        children: [{ id: "auth-menu-config", name: "角色菜单管理", path: "/role/menu" }],
      },
      { id: "auth-role", name: "权限列表", path: "/role/list" },
    ],
  },
];

export default navRouter;

/**
 * SSR 路由配置
 *
 * 添加新的 SSR 页面只需要：
 * 1. 在 src/pages/ 下新建页面组件
 * 2. 在本文件中添加一行路由配置
 *
 * 路由、导航、SSR 入口、server.ts 全部自动同步
 */
import type { ComponentType } from "react";

import { SSRDemo } from "@/pages/ssr-demo";
import { SSRPerformance } from "@/pages/ssr-demo/performance";
import { SSRDataFetch } from "@/pages/ssr-demo/data-fetch";

export interface SSRRouteItem {
  /** 完整路由路径 */
  path: string;
  /** 页面组件 */
  component: ComponentType;
  /** 页面标题（用于导航） */
  title: string;
}

export const ssrRoutes: SSRRouteItem[] = [
  { path: "/ssr", component: SSRDemo, title: "基础演示" },
  { path: "/ssr/performance", component: SSRPerformance, title: "性能测试" },
  { path: "/ssr/data-fetch", component: SSRDataFetch, title: "数据获取" },
];

/** 所有 SSR 路径集合（供 server.ts 判断是否走 SSR） */
export const ssrPaths: string[] = ssrRoutes.map(r => r.path);

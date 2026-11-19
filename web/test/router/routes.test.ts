import { describe, expect, it } from "vitest";

import { navRouter } from "@/router/section/nav-router";
import ROUTER_PATH from "@/router/section/router-path";
import { spaPaths, spaRoutes } from "@/router/section/spa-routes";

const flattenNavPaths = (
  nodes: typeof navRouter,
): string[] => nodes.flatMap(node => [node.path, ...flattenNavPaths(node.children ?? [])]);

describe("route configuration", () => {
  it("spaPaths 始终由 spaRoutes 派生且路径唯一", () => {
    expect(spaPaths).toEqual(spaRoutes.map(route => route.path));
    expect(new Set(spaPaths).size).toBe(spaPaths.length);
  });

  it("每条 SPA 路由都有标题和绝对路径", () => {
    for (const route of spaRoutes) {
      expect(route.title.trim()).not.toBe("");
      expect(route.path.startsWith("/")).toBe(true);
    }
  });

  it("路由常量中的主要页面都存在于 SPA 路径表", () => {
    const expectedPaths = [
      ROUTER_PATH.home,
      ROUTER_PATH.login,
      ROUTER_PATH.data,
      ROUTER_PATH.ws,
      ROUTER_PATH.user.root,
      ROUTER_PATH.user.operation,
      ROUTER_PATH.subscribe.home,
      ROUTER_PATH.subscribe.list,
      ROUTER_PATH.subscribe.detail,
      ROUTER_PATH.role.list,
      ROUTER_PATH.role.menu,
    ];

    expect(spaPaths).toEqual(expect.arrayContaining(expectedPaths));
  });

  it("导航中的可访问路径都已登记在 SPA 或 SSR 路由中", () => {
    const registeredPaths = new Set([...spaPaths, "/ssr", "/ssr/performance", "/ssr/data-fetch"]);
    const accessibleNavPaths = flattenNavPaths(navRouter).filter(Boolean);

    for (const path of accessibleNavPaths) {
      expect(registeredPaths.has(path), `未登记的导航路径: ${path}`).toBe(true);
    }
  });

  it("导航节点 id 唯一", () => {
    const flattenIds = (nodes: typeof navRouter): string[] =>
      nodes.flatMap(node => [node.id, ...flattenIds(node.children ?? [])]);
    const ids = flattenIds(navRouter);

    expect(new Set(ids).size).toBe(ids.length);
  });
});

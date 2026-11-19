import { describe, expect, it } from "vitest";

import { buildMenuTree, transformNode } from "@/utils/treeTransform";

import type { NavRouter } from "@/router";
import type { MenuItemType } from "@/types/menu-router";

const menuItem = (overrides: Partial<MenuItemType>): MenuItemType => ({
  id: "1",
  name: "菜单",
  link_url: "/menu",
  menu_code: "menu",
  parent_id: "0",
  node_type: 1,
  icon_url: "",
  level: 1,
  path: "",
  is_delete: 0,
  ...overrides,
});

describe("treeTransform", () => {
  it("将路由配置递归转换为 Ant Design Tree 数据", () => {
    const routes: NavRouter[] = [
      { id: "home", name: "首页", path: "/" },
      {
        id: "admin",
        name: "管理",
        path: "",
        children: [
          { id: "users", name: "用户", path: "/users" },
          {
            id: "settings",
            name: "设置",
            path: "",
            children: [{ id: "security", name: "安全", path: "/security" }],
          },
        ],
      },
    ];

    expect(transformNode([], routes)).toEqual([
      { key: "/", title: "首页" },
      {
        key: "menu-管理",
        title: "管理",
        children: [
          { key: "/users", title: "用户" },
          {
            key: "menu-设置",
            title: "设置",
            children: [{ key: "/security", title: "安全" }],
          },
        ],
      },
    ]);
  });

  it("空路由数组返回空树", () => {
    expect(transformNode([], [])).toEqual([]);
  });

  it("将无序的扁平菜单构建为多级树", () => {
    const flatData = [
      menuItem({ id: "3", name: "孙菜单", menu_code: "grandchild", parent_id: "2", level: 3 }),
      menuItem({ id: "1", name: "根菜单", menu_code: "root", parent_id: "0", level: 1 }),
      menuItem({ id: "2", name: "子菜单", menu_code: "child", parent_id: "1", level: 2 }),
      menuItem({ id: "4", name: "第二根菜单", menu_code: "root-2", parent_id: "0", level: 1 }),
    ];

    expect(buildMenuTree(flatData)).toEqual([
      {
        ...flatData[1],
        key: "root",
        title: "根菜单",
        children: [
          {
            ...flatData[2],
            key: "child",
            title: "子菜单",
            children: [
              {
                ...flatData[0],
                key: "grandchild",
                title: "孙菜单",
              },
            ],
          },
        ],
      },
      {
        ...flatData[3],
        key: "root-2",
        title: "第二根菜单",
      },
    ]);
  });

  it("支持 parent_id 为 null 的根节点，并忽略父节点不存在的孤儿节点", () => {
    const root = menuItem({ id: "root", parent_id: null as unknown as string });
    const orphan = menuItem({ id: "orphan", parent_id: "missing" });

    expect(buildMenuTree([root, orphan])).toEqual([
      { ...root, key: root.menu_code, title: root.name },
    ]);
  });
});

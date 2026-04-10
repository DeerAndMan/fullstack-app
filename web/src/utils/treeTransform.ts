import { navRouter } from "@/router";

import type { MenuItemType, TreeMenuItemType } from "@/types/menu-router";
import type { TreeDataNode } from "antd/lib";
import type { NavRouter } from "@/router";

/**
 * 转换路由数据为树结构
 * @param selectedList 选中的列表
 * @param routeData 全部路由数据
 * @returns
 */
export const transformNode = (
  selectedList: MenuItemType[] = [],
  routeData: NavRouter[] = navRouter
): TreeDataNode[] => {
  if (!navRouter.length) return [];
  const selectedData: React.Key[] = [];

  const transform = (node: NavRouter): TreeDataNode => {
    const key = (node.path === "" ? `menu-${node.name}` : node.path) as React.Key;
    const findItem = selectedList.find(s => s.link_url === key && s.menu_code === key);
    if (findItem) {
      selectedData.push(key);
    }

    const menuItem: TreeDataNode = { key, title: node.name };
    if (node.children && node.children.length > 0) {
      return { ...menuItem, children: node.children.map(child => transform(child)) };
    }
    return menuItem;
  };

  return routeData.map(node => transform(node));
};

/**
 * 构建菜单树
 * @param flatData 菜单数据
 * @returns
 */
export const buildMenuTree = (flatData: MenuItemType[]) => {
  const idMap: Record<string, TreeMenuItemType> = {};
  const rootNodes: TreeMenuItemType[] = [];

  flatData.forEach(node => {
    idMap[node.id] = { ...node, key: node.menu_code, title: node.name, children: [] };
  });

  flatData.forEach(node => {
    const currentNode = idMap[node.id];
    if (node.parent_id === "0" || node.parent_id === null) {
      rootNodes.push(currentNode);
    } else {
      const parentNode = idMap[node.parent_id];
      if (parentNode && parentNode.children) {
        parentNode.children.push(currentNode);
      }
    }
  });

  removeEmptyChildren(rootNodes);

  return rootNodes;
};

/**
 * 移除空的子节点
 * @param nodes 节点列表
 */
const removeEmptyChildren = (nodes: TreeMenuItemType[]) => {
  nodes.forEach(node => {
    if (node.children && node.children.length > 0) {
      removeEmptyChildren(node.children);
    }
    if (node.children && node.children.length === 0) {
      delete node.children;
    }
  });
};

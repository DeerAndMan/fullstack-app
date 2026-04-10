import type { TreeDataNode } from "antd";

export type MenuItemType = {
  id: string; // 主键（对应Go的int64）
  name: string; // 名称（对应Go的varchar(100)，非空）
  link_url: string; // 页面对应的地址（对应Go的varchar(500)，非空）
  menu_code: string; // 菜单编码、别名（对应Go的varchar(100)，非空）
  parent_id: string; // 父节点（允许为NULL，对应Go的bigint）
  node_type: number; // 父节点类型：1 页面、2 按钮（允许为NULL，对应Go的tinyint）
  icon_url: string; // 图标地址（允许为NULL，对应Go的varchar(255)）
  level: number; // 层次（对应Go的int，非空）
  path: string; // 树 ID 的路径（允许为NULL，对应Go的varchar(2500)）
  is_delete: number; // 是否删除：1 已删除、0 未删除（对应Go的tinyint，非空）
};

export type TreeMenuItemType = MenuItemType & {
  key: string;
  title: string;
  children?: TreeMenuItemType[];
};

export type MenuRouterType = {} & TreeDataNode;

// 角色路由
export type RoleRoutingType = {
  roleId: number; // 角色ID（对应Go的bigint，非空）
  menuIds: string[]; // 菜单ID列表（对应Go的json数组，非空）
};

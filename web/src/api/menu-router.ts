import { request, apiControl } from "@/api";

import type { PromiseResponseData } from "@/api";
import type { MenuItemType, MenuRouterType, RoleRoutingType } from "@/types/menu-router";

export const getMenuRouterList = (): PromiseResponseData<MenuItemType[]> =>
  request.get(apiControl.menu.list);

export const addMenuRouter = (params: MenuRouterType[]): PromiseResponseData =>
  request.post(apiControl.menu.add, params);

export const addRoleRouting = (params: RoleRoutingType): PromiseResponseData =>
  request.post(apiControl.menu.roleRouting, params);

export const getRoleRoutingByRoleId = (roleId: number): PromiseResponseData<MenuItemType[]> =>
  request.get(apiControl.menu.roleListByRoleId + roleId);

import { request } from "../";
import { apiPathsV1 } from "../paths/v1";

import type { PromiseResponseData } from "../";
import type { MenuItemType, MenuRouterType, RoleRoutingType } from "@/types/menu-router";

export const getMenuRouterList = (): PromiseResponseData<MenuItemType[]> =>
  request.get(apiPathsV1.menu.list);

export const addMenuRouter = (params: MenuRouterType[]): PromiseResponseData =>
  request.post(apiPathsV1.menu.add, params);

export const addRoleRouting = (params: RoleRoutingType): PromiseResponseData =>
  request.post(apiPathsV1.menu.roleBinding, params);

export const getRoleRoutingByRoleId = (roleId: number): PromiseResponseData<MenuItemType[]> =>
  request.get(`${apiPathsV1.menu.roleListByRoleId}/${roleId}`);

import { apiControl, request } from "@/api";

import type { AxiosRequestHeaders } from "axios";
import type { PartialCustomRequestConfig, PromiseResponseData } from "@/api";
import type { Account, AddAccount, Role } from "@/types/user";
import type { MenuItemType } from "@/types/menu-router";

export type LoginParams = {
  username: string;
  password: string;
};

export const login = (
  params: LoginParams,
  other: PartialCustomRequestConfig
): PromiseResponseData<{ token: string; user: Account; role: Role; menuRoles: MenuItemType[] }> =>
  request.post(`${apiControl.user.login}`, params, { ...other });

export const logout = (): PromiseResponseData => request.get(`${apiControl.user.admin}/logout`);

export const getAllUser = (): PromiseResponseData<Account[]> => request.get(apiControl.user.admin);
export const uploadAvatar = (params: FormData): PromiseResponseData =>
  request.put(apiControl.user.admin + "/upload-avatar", params, {
    headers: {
      "Content-Type": "multipart/form-data",
    } as AxiosRequestHeaders,
  });

export const deleteUser = (id: number): PromiseResponseData =>
  request.delete(`/admin/user/del`, { params: { id } });

export const addUser = (params: AddAccount): PromiseResponseData =>
  request.post("/admin/user/add", params);

export const updateUser = (params: AddAccount & { id: number }): PromiseResponseData =>
  request.put("/admin/user/update", params);

export const updateUserRole = (params: {
  user_id: number;
  role_id: number;
}): PromiseResponseData<Role> => request.put("/admin/user/role", params);

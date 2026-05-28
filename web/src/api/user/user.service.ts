import { request } from "../";
import { apiPathsV1 } from "../paths/v1";

import type { AxiosRequestHeaders } from "axios";
import type { PageData, PartialCustomRequestConfig, PromiseResponseData } from "../";
import type { Account, AddAccount, Role } from "@/types/user";
import type { MenuItemType } from "@/types/menu-router";

export type LoginParams = {
  name: string;
  password: string;
};

export type LoginResponse = {
  token: { access_token: string; refresh_token: string; expires_at: number };
  user: Account;
  role: Role;
  menuRoles: MenuItemType[];
};

export const login = (
  params: LoginParams,
  other: PartialCustomRequestConfig
): PromiseResponseData<LoginResponse> =>
  request.post(apiPathsV1.auth.login, params, { ...other });

export const logout = (): PromiseResponseData => request.post(apiPathsV1.auth.logout);

export const getAllUser = (): PromiseResponseData<PageData<Account>> => request.get(apiPathsV1.user.list);

export const getProfile = (): PromiseResponseData<Account> => request.get(apiPathsV1.user.profile);

export const uploadAvatar = (params: FormData): PromiseResponseData =>
  request.post(apiPathsV1.upload, params, {
    headers: {
      "Content-Type": "multipart/form-data",
    } as AxiosRequestHeaders,
  });

export const deleteUser = (id: number): PromiseResponseData =>
  request.delete(`${apiPathsV1.user.detail}/${id}`);

export const addUser = (params: AddAccount): PromiseResponseData =>
  request.post(apiPathsV1.user.list, params);

export const updateUser = (params: AddAccount & { id: number }): PromiseResponseData =>
  request.put(`${apiPathsV1.user.detail}/${params.id}`, params);

export const updateUserRole = (params: {
  user_id: number;
  role_id: number;
}): PromiseResponseData<Role> => request.put(`${apiPathsV1.user.detail}/${params.user_id}`, params);

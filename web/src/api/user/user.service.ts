import { apiControl, request } from "../";

import type { AxiosRequestHeaders } from "axios";
import type { PartialCustomRequestConfig, PromiseResponseData } from "../";
import type { Account, AddAccount, Role } from "@/types/user";

export type LoginParams = {
  name: string;
  password: string;
};

export const login = (
  params: LoginParams,
  other: PartialCustomRequestConfig
): PromiseResponseData<{ token: { access_token: string; refresh_token: string; expires_at: string }; user: Account }> =>
  request.post(apiControl.auth.login, params, { ...other });

export const logout = (): PromiseResponseData => request.post(apiControl.auth.logout);

export const getAllUser = (): PromiseResponseData<Account[]> => request.get(apiControl.user.list);

export const getProfile = (): PromiseResponseData<Account> => request.get(apiControl.user.profile);

export const uploadAvatar = (params: FormData): PromiseResponseData =>
  request.post(apiControl.upload, params, {
    headers: {
      "Content-Type": "multipart/form-data",
    } as AxiosRequestHeaders,
  });

export const deleteUser = (id: number): PromiseResponseData =>
  request.delete(`${apiControl.user.detail}/${id}`);

export const addUser = (params: AddAccount): PromiseResponseData =>
  request.post(apiControl.user.list, params);

export const updateUser = (params: AddAccount & { id: number }): PromiseResponseData =>
  request.put(`${apiControl.user.detail}/${params.id}`, params);

export const updateUserRole = (params: {
  user_id: number;
  role_id: number;
}): PromiseResponseData<Role> => request.put(`${apiControl.user.detail}/${params.user_id}`, params);

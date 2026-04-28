import { useQuery, useQueryClient, useMutation } from "@tanstack/react-query";

import * as userApi from "./user.service";

import type { CallbackFunction } from "@/types/constants";
import type { ResponseData } from "../";
import type { Account, Role } from "@/types/user";

export const userQuery = (enabled = true) => {
  const queryClient = useQueryClient();

  const queryList = useQuery({
    queryKey: ["user", "list"],
    queryFn: () => userApi.getAllUser(),
    enabled,
  });

  const refreshQuery = () => {
    queryClient.invalidateQueries({ queryKey: ["user", "list"], exact: true });
  };

  return { queryList, refreshQuery };
};

export const getQueryUserList = () => {
  const queryClient = useQueryClient();
  const queryList = queryClient.getQueryData<ResponseData<Account[]>>(["user", "list"]);
  return queryList?.data || [];
};

export const userRoleMutation = (callback?: CallbackFunction<Role>) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { user_id: number; role_id: number }) => userApi.updateUserRole(params),
    onSuccess: data => {
      if (data.code !== 0) {
        callback?.errorCb?.({ msg: data.msg });
        return;
      }
      callback?.successCb?.(data.data);
      queryClient.invalidateQueries({ queryKey: ["user", "list"], exact: true });
    },
    onError: err => {
      callback?.errorCb?.({ msg: err.message });
    },
    onSettled: () => {
      callback?.finallyCb?.();
    },
  });
};

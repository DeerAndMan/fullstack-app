import { useQuery, useQueryClient, useMutation } from "@tanstack/react-query";

import { useGlobalStore } from "@/stores/global";
import { getMenuRouterList, addRoleRouting, getRoleRoutingByRoleId } from "./menu.service";

import type { RoleRoutingType } from "@/types/menu-router";

const queryKey = ["menu", "router", "list"];

export const menuListQuery = (enabled = true) => {
  const queryClient = useQueryClient();

  const queryList = useQuery({
    queryKey,
    queryFn: () => getMenuRouterList().then(res => (res.code === 0 ? res.data : [])),
    enabled,
  });

  const refresh = () => {
    queryClient.invalidateQueries({ queryKey });
  };

  return { queryList, refresh };
};

export const roleRoutingMutation = () => {
  const messageApi = useGlobalStore(s => s.messageApi);

  return useMutation({
    mutationKey: ["menu", "router", "roleRouting"],
    mutationFn: (params: RoleRoutingType) => addRoleRouting(params),
    onSuccess: data => {
      if (data.code !== 0) {
        messageApi?.error(data.message || "添加角色路由失败!!!!");
        return;
      }
    },
    onError: () => {},
  });
};

export const getRoleRoutingByRoleIdQuery = (roleId: number) =>
  useQuery({
    queryKey: ["menu", "router", "roleRoutingByRoleId", roleId],
    queryFn: () => getRoleRoutingByRoleId(roleId),
    enabled: !!roleId,
  });

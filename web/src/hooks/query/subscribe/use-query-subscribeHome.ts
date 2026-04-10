import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import z from "zod";

import { apiControl, RequestPost, RequestPut, RequestSchema } from "@/api";
import { useGlobalStore } from "@/stores/global";

import { SubscribeListSchema } from "@/types/xq/subscribe/home";

const queryKey = ["subscribe", "home", "list"];

export const useQuerySubscribeHome = () => {
  const queryClient = useQueryClient();
  const messageApi = useGlobalStore(s => s.messageApi);

  const queryList = useQuery({
    queryKey: [...queryKey],
    queryFn: async () => {
      const res = await RequestSchema({
        method: "GET",
        url: apiControl.xq.subscribe.list,
        schema: SubscribeListSchema,
      });
      const newData = res.data.map(d => ({ ...d, key: d.id }));

      return { ...res, data: newData };
    },
    enabled: true,
  });

  const toggleMutation = useMutation({
    mutationFn: (params: { user_id: string; enabled: boolean }) =>
      RequestPut(
        `${apiControl.xq.subscribe.toggle}/${params.user_id}`,
        { enabled: params.enabled },
        { schema: z.object({}).nullable() }
      ),
    onError: (error: Error) => {
      messageApi?.error(error.message || "状态切换失败");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey, exact: true });
    },
  });

  const addMutation = useMutation({
    mutationFn: (params: { user_id: number; description?: string }) =>
      RequestPost(apiControl.xq.subscribe.add, params, {
        schema: z.object({}).nullable(),
      }),
    onError: (error: Error) => {
      messageApi?.error(error.message || "添加订阅失败");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey, exact: true });
    },
  });

  const updateMutation = useMutation({
    mutationFn: (params: { id: string; user_id: string; former_name: string }) =>
      RequestPut(
        `${apiControl.xq.subscribe.formerName}/${params.id}/${params.user_id}`,
        { former_name: params.former_name },
        { schema: z.object({}).nullable() }
      ),
    onError: (error: Error) => {
      messageApi?.error(error.message || "添加曾用名失败");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey, exact: true });
    },
  });

  return { queryList, toggleMutation, addMutation, updateMutation };
};

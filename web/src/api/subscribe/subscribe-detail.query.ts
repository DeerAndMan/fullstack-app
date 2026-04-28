import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import z from "zod";

import { apiControl, RequestSchema } from "../";
import { SubscribeItemSchema, ThemeContentItemSchema } from "@/types/xq/subscribe/home";

const DetailTableSchema = z.object({
  pageNumber: z.number(),
  pageSize: z.number(),
  totalPage: z.number(),
  totalCount: z.number(),
  list: z.array(ThemeContentItemSchema),
});

export const useQuerySubscribeDetail = (id: string, userId: string) => {
  const [pageNumber, setPageNumber] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const queryDetail = useQuery({
    queryKey: ["subscribe", "detail", id, userId],
    queryFn: async () => {
      const res = await RequestSchema({
        method: "GET",
        url: `${apiControl.xq.subscribe.detail}/${id}/${userId}`,
        schema: SubscribeItemSchema,
      });
      return res;
    },
    enabled: !!id && !!userId,
  });

  const queryTable = useQuery({
    queryKey: ["subscribe", "detail-table", userId, pageNumber, pageSize],
    queryFn: async () => {
      const res = await RequestSchema({
        method: "GET",
        url: `${apiControl.xq.subscribe.detailTable}/${userId}?pageNumber=${pageNumber}&pageSize=${pageSize}`,
        schema: DetailTableSchema,
      });
      return res;
    },
    enabled: !!userId,
  });

  return { queryDetail, queryTable, pageNumber, setPageNumber, pageSize, setPageSize };
};

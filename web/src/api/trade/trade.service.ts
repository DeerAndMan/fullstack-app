import z from "zod";
import { apiControl, request, RequestSchema } from "../";
import { TradeItemSchema } from "@/types/schema";

import type { TradeItem } from "@/pages/trade/type";
import type { PromiseResponseData } from "../";
import type { TradeParams } from "@/types/schema";

// 使用类型化的请求函数，返回类型从 schema 自动推断
export const getTrade = (params: TradeParams) =>
  RequestSchema({
    method: "POST",
    url: apiControl.trade.index,
    data: params,
    schema: z.array(TradeItemSchema),
  });

export const getSummary = (params: TradeParams): PromiseResponseData<TradeItem> =>
  request.post(apiControl.trade.summary, params);

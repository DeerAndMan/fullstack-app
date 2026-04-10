/**
 * 修复后的 Schema 定义
 * 基于实际数据结构进行调整
 */

import { z } from "zod";

// 基础响应类型 - 更宽松的定义
export const BaseResponseSchema = z.object({
  data: z.any(),
  code: z.number(),
  msg: z.string().optional(),
});

// 能源项目 Schema - 基于实际数据调整
export const EnergyItemSchema = z.object({
  id: z.number(),
  datetime: z.string(),
  Zqmc: z.string(),
  Cbjgex: z.string(),
  Cbjg: z.string(),
  Ckcb: z.string(),
  Ckcbj: z.string(),
  Ckyk: z.string(),
  Ykbl: z.string(),
  Dryk: z.string(),
  Drykbl: z.string(),
  Cwbl: z.string(),
  Djsl: z.string(),
  Dqcb: z.string(),
  Gddm: z.string(),
  Gfmcdj: z.string(),
  Gfmrjd: z.string(),
  Gfssmmce: z.string(),
  Gfye: z.string(),
  Jgbm: z.string(),
  Khdm: z.string(),
  Ksssl: z.string(),
  Kysl: z.string(),
  Ljyk: z.string(),
  Market: z.string(),
  Mrssc: z.string(),
  Sssl: z.string(),
  Szjsbs: z.string(),
  Zjzh: z.string(),
  Zqdm: z.string(),
  Zqlx: z.string(),
  Zqlxmc: z.string(),
  Zqsl: z.string(),
  Ztmc: z.string(),
  Ztmr: z.string(),
  Zxjg: z.string(),
  Zxsz: z.string(),
  Bz: z.string(),
});

// 交易项目 Schema - 基于实际数据调整
export const TradeItemSchema = z.object({
  id: z.number(),
  date: z.string(),
  drhz: z.string(),
  dryk: z.string(),
  zxsz: z.string(),
  zzc: z.string(),
  RMBZzc: z.string(),
  num: z.number(),
  zsz: z.string(),
  ccyk: z.string(),
  stocks: z.string(),
  zjye: z.string(),
  positions: z.array(EnergyItemSchema),
  djzj: z.number(),
  kqzj: z.number(),
  ljyk: z.number(),
  kyzj: z.number(),
  money_type: z.string(),
  totalsecMKval: z.string(),
});

// 更宽松的交易项目 Schema - 允许可选字段
export const FlexibleTradeItemSchema = TradeItemSchema.partial().extend({
  id: z.number(),
  positions: z.array(EnergyItemSchema.partial()),
});

// 交易参数 Schema
export const TradeParamsSchema = z.object({
  startTime: z.string().min(1, "开始时间不能为空"),
  endTime: z.string().min(1, "结束时间不能为空"),
});

// 响应 Schema - 直接返回数据
export const TradeListResponseSchema = z.array(TradeItemSchema);
export const TradeSummaryResponseSchema = TradeItemSchema;

// 包装的响应 Schema
export const WrappedTradeListResponseSchema = BaseResponseSchema.extend({
  data: z.array(TradeItemSchema),
});

export const WrappedTradeSummaryResponseSchema = BaseResponseSchema.extend({
  data: TradeItemSchema,
});

// 导出类型
export type TradeParams = z.infer<typeof TradeParamsSchema>;
export type EnergyItem = z.infer<typeof EnergyItemSchema>;
export type TradeItem = z.infer<typeof TradeItemSchema>;
export type FlexibleTradeItem = z.infer<typeof FlexibleTradeItemSchema>;
export type TradeListResponse = z.infer<typeof TradeListResponseSchema>;
export type TradeSummaryResponse = z.infer<typeof TradeSummaryResponseSchema>;
export type BaseResponse = z.infer<typeof BaseResponseSchema>;

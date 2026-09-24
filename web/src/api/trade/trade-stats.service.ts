import { RequestGet, RequestPost } from "../request-schema";
import { apiPathsV1 } from "../paths/v1";
import {
  StatsDateRangeSchema,
  StatsQueryResponseSchema,
  SyncDailyResponseSchema,
  type StatsQueryParams,
  type SyncDailyParams,
} from "@/types/schema";

/**
 * 查询交易统计数据
 * @param params 查询参数
 * @returns 统计数据列表和总体汇总
 */
export const queryTradeStats = (params: StatsQueryParams) =>
  RequestPost(apiPathsV1.trade.stats.query, params, { schema: StatsQueryResponseSchema });

/**
 * 同步日终快照（从 summary 表提取到 trade_daily_summary）
 * @param params 同步参数（可选，缺省同步今天）
 * @returns 同步结果
 */
export const syncDailySnapshots = (params?: SyncDailyParams) =>
  RequestPost(apiPathsV1.trade.stats.sync, params ?? {}, { schema: SyncDailyResponseSchema });

/**
 * 查询日终快照的可用日期范围
 * @returns 最早 / 最晚交易日与交易日总数，供前端把默认区间对齐到真实数据
 */
export const fetchStatsDateRange = () =>
  RequestGet(apiPathsV1.trade.stats.range, { schema: StatsDateRangeSchema });

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";

import { fetchStatsDateRange, queryTradeStats, syncDailySnapshots } from "./trade-stats.service";

import type { Dayjs } from "dayjs";

const dateFormat = "YYYY-MM-DD";
const queryKey = ["trade", "stats"];
const rangeKey = [...queryKey, "range"];

/**
 * 页面上可选的汇总粒度。
 *
 * 注意：这里的粒度只决定「默认看多长的时间窗口」，不决定聚合方式——
 * 所有粒度都按天展示每一天的数据（后端固定走 groupBy=day），
 * 「周度」= 这一周的每天，「年度」= 这一年的每天，「整体」= 全部历史的每天。
 */
export type GroupByType = "week" | "month" | "half_year" | "year" | "custom";

/** 日终快照的真实数据边界 */
type DataBounds = { min: Dayjs; max: Dayjs };

/** 粒度选项，供页面下拉/分段器使用 */
export const groupByOptions: { label: string; value: GroupByType }[] = [
  { label: "周度", value: "week" },
  { label: "月度", value: "month" },
  { label: "半年度", value: "half_year" },
  { label: "年度", value: "year" },
  { label: "整体", value: "custom" },
];

/**
 * 按粒度计算默认查询区间：以日终快照的最晚交易日为锚点，
 * 取它所在的自然周 / 月 / 半年 / 年，起点不早于最早交易日。
 * 「整体」直接覆盖全部历史。
 */
const defaultRangeOf = (groupBy: GroupByType, bounds: DataBounds): [Dayjs, Dayjs] => {
  const end = bounds.max.endOf("day");

  const start = (() => {
    switch (groupBy) {
      case "week":
        return end.startOf("week");
      case "month":
        return end.startOf("month");
      // 半年度：1-6 月归上半年，7-12 月归下半年
      case "half_year":
        return end.month() < 6 ? end.startOf("year") : end.month(6).startOf("month");
      case "year":
        return end.startOf("year");
      // 整体：拉满全部历史
      default:
        return bounds.min;
    }
  })();

  // 窗口起点早于已有数据时，退回到最早交易日
  return [start.isBefore(bounds.min) ? bounds.min.startOf("day") : start.startOf("day"), end];
};

/**
 * 交易统计查询 hook。
 * 维护汇总粒度与自定义日期区间，切换粒度时自动套用对应的默认区间；
 * 默认区间对齐日终快照的真实数据范围，而不是相对今天硬编码。
 */
export const tradeStatsQuery = (enabled = true) => {
  const queryClient = useQueryClient();
  const [groupBy, setGroupBy] = useState<GroupByType>("month");
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs] | null>(null);

  // 先拿到日终快照的日期范围，再据此决定默认查询区间
  const rangeQuery = useQuery({
    queryKey: rangeKey,
    queryFn: fetchStatsDateRange,
    enabled,
    staleTime: 5 * 60 * 1000,
  });

  const range = rangeQuery.data?.data;
  const bounds: DataBounds | null =
    range && range.minDate && range.maxDate
      ? { min: dayjs(range.minDate), max: dayjs(range.maxDate) }
      : null;

  // 数据范围到手后初始化一次默认区间（用户手动改过则不再覆盖）
  useEffect(() => {
    if (bounds && !dateRange) {
      setDateRange(defaultRangeOf(groupBy, bounds));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [range?.minDate, range?.maxDate]);

  const statsQuery = useQuery({
    queryKey: [
      ...queryKey,
      groupBy,
      dateRange?.[0].format(dateFormat) ?? "",
      dateRange?.[1].format(dateFormat) ?? "",
    ],
    queryFn: () =>
      // 所有粒度都按天出数据，粒度只影响上面的默认区间
      queryTradeStats({
        groupBy: "day",
        startDate: dateRange![0].format(dateFormat),
        endDate: dateRange![1].format(dateFormat),
      }),
    enabled: enabled && !!dateRange,
  });

  // 同步日终快照，成功后刷新日期范围与统计结果
  const syncMutation = useMutation({
    mutationFn: (params?: { startDate?: string; endDate?: string }) => syncDailySnapshots(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });

  // 切换粒度时同步套用该粒度的默认区间
  const changeGroupBy = (next: GroupByType) => {
    setGroupBy(next);
    if (bounds) {
      setDateRange(defaultRangeOf(next, bounds));
    }
  };

  const refresh = () => {
    queryClient.invalidateQueries({ queryKey });
  };

  const stateOperations = {
    groupBy,
    setGroupBy: changeGroupBy,
    dateRange,
    setDateRange: (next: [Dayjs, Dayjs]) => setDateRange(next),
    /** 数据边界，供日期选择器限制可选范围 */
    bounds,
  };
  const operations = { refresh, sync: syncMutation.mutate, syncing: syncMutation.isPending };

  return { statsQuery, rangeQuery, stateOperations, operations };
};

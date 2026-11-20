import { useQuery } from "@tanstack/react-query";
import { z } from "zod";

import { RequestGet } from "@/api/request-schema";
import { apiPathsV1 } from "@/api/paths/v1";

const rowSchema = z.object({
  shanghaiChange: z.number().optional(), shenzhenChange: z.number().optional(),
  peTtm: z.number().optional(),
  pbMrq: z.number().optional(),
  psTtm: z.number().optional(),
  pcfOcfTtm: z.number().optional(),
  quoteSource: z.string().optional(), scoreUrl: z.string().url().optional(),
  advancers: z.number().optional(), decliners: z.number().optional(), sentiment: z.number().optional(),
  code: z.string(),
  symbol: z.string().optional(),
  name: z.string(),
  date: z.string().optional(),
  detail: z.string().optional(),
  url: z.string().url().optional(),
  price: z.number().optional(),
  open: z.number().optional(),
  high: z.number().optional(),
  low: z.number().optional(),
  volume: z.number().optional(),
  average: z.number().optional(),
  change: z.number().optional(),
  amount: z.number().optional(),
  net: z.number().optional(),
  count: z.number().optional(),
  priority: z.number().optional(),
  previous: z.string().optional(),
  forecast: z.string().optional(),
  actual: z.string().optional(),
});

const datasetSchema = z.object({
  source: z.string(),
  sourceUrl: z.string().url().optional(),
  fetchedAt: z.string(),
  status: z.enum(["ready", "stale", "unavailable"]),
  message: z.string().optional(),
  rows: z.array(rowSchema),
  total: z.number().optional(),
  adjustment: z.enum(["none", "qfq"]).optional(),
});

const conceptReasonSchema = z.object({
  status: z.enum(["ready", "missing", "unavailable"]),
  kind: z.string().optional(),
  text: z.string().optional(),
  sources: z.array(z.object({ title: z.string(), url: z.url({ protocol: /^https$/ }) })).optional(),
});

export type MarketRow = z.infer<typeof rowSchema>;

export function isStockSymbol(symbol?: string): symbol is string {
  return Boolean(symbol && /^(?:sh6[0-9]{5}|sz(?:00|30)[0-9]{4}|bj[489][0-9]{5})$/.test(symbol));
}

export const marketFeeds = {
  "market-history": "https://www.kaipanla.com/omain/",
  "market-history-ths": "https://stockpage.10jqka.com.cn/1A0001/",
  "market-history-sina": "https://finance.sina.com.cn/realstock/company/sh000001/nc.shtml",
  "market-history-eastmoney": "https://quote.eastmoney.com/sh000001.html",
  indices: "https://gu.qq.com/",
  industry: "https://stocks.sina.cn/sh/fieldranking?type=sina&vt=4",
  concept: "https://finance.sina.com.cn/stock/",
  "concept-eastmoney": "https://quote.eastmoney.com/center/boardlist.html#concept_board",
  "concept-ths": "https://data.10jqka.com.cn/funds/gnzjl/",
  "industry-eastmoney": "https://quote.eastmoney.com/center/boardlist.html#industry_board",
  "industry-ths": "https://q.10jqka.com.cn/thshy/",
  "concept-members": "https://finance.sina.com.cn/stock/",
  "concept-news": "https://wallstreetcn.com/search",
  "stock-quote": "https://gu.qq.com/",
  "stock-minute": "https://gu.qq.com/",
  "stock-five-day": "https://gu.qq.com/",
  "stock-daily": "https://gu.qq.com/",
  "stock-weekly": "https://gu.qq.com/",
  "stock-monthly": "https://gu.qq.com/",
  "stock-valuation": "https://data.eastmoney.com/gzfx/",
  "fund-flow": "https://data.eastmoney.com/bkzj/",
  "fund-flow-eastmoney": "https://data.eastmoney.com/bkzj/",
  "fund-flow-sina": "https://vip.stock.finance.sina.com.cn/moneyflow/#blocktol_zjh",
  "fund-flow-ths": "https://q.10jqka.com.cn/thshy/",
  "region-fund-flow": "https://data.eastmoney.com/bkzj/dy.html",
  "region-fund-flow-sina": "https://vip.stock.finance.sina.com.cn/mkt/#diyu",
  "region-fund-flow-ths": "https://q.10jqka.com.cn/dy/",
  "board-members": "https://finance.sina.com.cn/stock/",
  "limit-up": "https://quote.eastmoney.com/ztb/ranklist.html",
  "limit-up-ths": "https://data.10jqka.com.cn/limit_up/",
  "limit-down": "https://quote.eastmoney.com/ztb/ranklist.html",
  "limit-down-ths": "https://data.10jqka.com.cn/limit_up/",
  "limit-down-xuangu": "https://xuangubao.com.cn/",
  "opened-limit": "https://quote.eastmoney.com/ztb/ranklist.html",
  "opened-limit-ths": "https://data.10jqka.com.cn/limit_up/",
  flash: "https://wallstreetcn.com/live",
  "flash-sina": "https://finance.sina.com.cn/7x24/",
  "flash-eastmoney": "https://kuaixun.eastmoney.com/",
  "flash-kaipanla": "https://www.kaipanla.com/page",
  "flash-ths": "https://stock.10jqka.com.cn/",
  calendar: "https://wallstreetcn.com/calendar",
  "calendar-eastmoney": "https://emdata.eastmoney.com/cjrl/index.html",
  "calendar-sina": "https://finance.sina.com.cn/stock/usstock/",
  "calendar-ths": "https://stock.10jqka.com.cn/fincalendar.shtml",
  leaderboard: "https://data.eastmoney.com/stock/lhb.html",
  "leaderboard-kaipanla": "https://www.kaipanla.com/omain/",
  "leaderboard-sina": "https://vip.stock.finance.sina.com.cn/q/go.php/vInvestConsult/kind/lhb/index.phtml",
} as const;

export type MarketFeed = keyof typeof marketFeeds;
type MarketParams = { date?: string; symbol?: string };

export function useMarketFeed(feed: MarketFeed, params?: MarketParams, enabled = true) {
  const flash = feed === "flash" || feed.startsWith("flash-");
  const valuation = feed === "stock-valuation";
  return useQuery({
    enabled,
    queryKey: ["market", feed, params],
    queryFn: async ({ signal }) =>
      (await RequestGet(`${apiPathsV1.market}/${feed}`, {
        params, signal, schema: datasetSchema, errorMsg: false, timeout: 12_000,
      })).data,
    staleTime: query => valuation && query.state.data?.status === "ready" ? 30 * 60_000 : 60_000,
    refetchInterval: flash || valuation ? false : 60_000,
    refetchOnMount: flash ? false : undefined,
    refetchOnWindowFocus: flash ? false : undefined,
    refetchOnReconnect: flash ? false : undefined,
    networkMode: flash ? "always" : undefined,
    retry: false,
  });
}

export type MarketQuery = ReturnType<typeof useMarketFeed>;

export function useConceptReason(concept: string, symbol: string) {
  return useQuery({
    queryKey: ["market-concept-reason", concept, symbol],
    queryFn: async ({ signal }) =>
      (await RequestGet(`${apiPathsV1.market}/concepts/${encodeURIComponent(concept)}/stocks/${encodeURIComponent(symbol)}/reason`, {
        signal, schema: conceptReasonSchema, errorMsg: false, timeout: 18_000,
      })).data,
    staleTime: query => query.state.data?.status === "unavailable" ? 0 : 60 * 60_000,
    gcTime: 60 * 60_000,
    retry: false,
  });
}

const calendarFormatter = new Intl.DateTimeFormat("sv-SE", {
  timeZone: "Asia/Shanghai", year: "numeric", month: "2-digit", day: "2-digit",
});
const updateFormatter = new Intl.DateTimeFormat("zh-CN", {
  timeZone: "Asia/Shanghai", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit",
});

export const localDate = () => calendarFormatter.format(new Date());
export const updateTime = (iso: string) => updateFormatter.format(new Date(iso));

export function marketSession(rows?: MarketRow[]) {
  const date = rows?.find(row => row.code === "sh000001")?.date?.slice(0, 10);
  return date && /^\d{4}-\d{2}-\d{2}$/.test(date) ? date : localDate();
}

const priceFormat = new Intl.NumberFormat("zh-CN", { minimumFractionDigits: 2, maximumFractionDigits: 3 });
export const priceText = (amount?: number) => amount === undefined ? "—" : priceFormat.format(amount);
export const moneyText = (amount?: number) => {
  if (amount === undefined) return "—";
  if (Math.abs(amount) >= 100_000_000) return `${(amount / 100_000_000).toFixed(2)} 亿`;
  return `${(amount / 10_000).toFixed(2)} 万`;
};
export const percentText = (value?: number) =>
  value === undefined ? "—" : `${value > 0 ? "+" : ""}${value.toFixed(2)}%`;

export function exportRows(label: string, headers: string[], values: (string | number | undefined)[][]) {
  const escape = (value: string | number | undefined) => {
    const literal = String(value ?? "");
    const safe = typeof value === "string" && /^\s*[=+@-]/.test(value) ? `'${literal}` : literal;
    return `"${safe.replaceAll('"', '""')}"`;
  };
  const contents = [headers, ...values].map(row => row.map(escape).join(",")).join("\r\n");
  const url = URL.createObjectURL(new Blob(["\uFEFF", contents], { type: "text/csv;charset=utf-8" }));
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `${label}-${localDate()}.csv`;
  anchor.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 1000);
}

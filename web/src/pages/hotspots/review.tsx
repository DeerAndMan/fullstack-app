import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Alert, Button, Card, Drawer, Empty, Input, Pagination, Select, Skeleton, Table, Tabs } from "antd";
import { useSearchParams } from "react-router-dom";
import { z } from "zod";
import { RequestGet } from "@/api/request-schema";
import styles from "./review.module.css";
import ReviewUpdate from "./review-update";
import ReviewTimeline from "./review-timeline";
import { reviewGuidance } from "./review-guidance";
import { PatternCard, SectorCard, StyleCard } from "./review-studies";

const daySchema = z.object({
  date: z.string(),
  previous: z.string(),
  batch: z.string(),
  rule: z.string(),
  release: z.string(),
  source: z.string(),
  windowStart: z.string(),
  total: z.number(),
  traded: z.number(),
  up: z.number(),
  highEligible: z.number(),
  highs: z.number(),
  csi: z.number(),
  eligible: z.number(),
  above: z.number(),
});
const reportSchema = z.object({ day: daySchema, highDelta: z.number().nullable() }).nullable();
const memberSchema = z.object({
  symbol: z.string(),
  name: z.string(),
  csi: z.boolean(),
  bars: z.number(),
  change: z.number().optional(),
  close: z.number().optional(),
  high: z.number().optional(),
  previousHigh: z.number().optional(),
  peakDate: z.string().optional(),
  newHigh: z.boolean().optional(),
  ma50: z.number().optional(),
  above: z.boolean().optional(),
  reason: z.string().optional(),
});
const listSchema = z.object({ rows: z.array(memberSchema), total: z.number() });
const labels = { up: "上涨名单", high: "一年新高名单", csi: "沪深 300 成份", excluded: "排除明细", all: "全部样本" };
type Metric = keyof typeof labels;
type Row = z.infer<typeof memberSchema>;
type Day = z.infer<typeof daySchema>;
const format = (v?: number) => (v === undefined ? "—" : v.toLocaleString("zh-CN", { maximumFractionDigits: 4 }));
const yes = (v?: boolean) => (v === undefined ? "—" : v ? "是" : "否");

function Members({ day, metric, close }: { day: Day; metric: Metric; close: () => void }) {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const list = useQuery({
    queryKey: ["review-members", day.batch, metric, search, page],
    queryFn: async ({ signal }) =>
      (
        await RequestGet("/api/v1/review/members", {
          params: { date: day.date, batch: day.batch, metric, page, search },
          schema: listSchema,
          signal,
        })
      ).data,
  });
  const base = [
    { title: "代码", dataIndex: "symbol" },
    { title: "名称", dataIndex: "name" },
    { title: "收盘价", dataIndex: "close", render: format },
  ];
  const price = [
    { title: "当日最高", dataIndex: "high", render: format },
    { title: "此前 249 根最高（当日基准）", dataIndex: "previousHigh", render: format },
    { title: "前高日期", dataIndex: "peakDate", render: (v?: string) => v || "—" },
  ];
  const columns = [
    ...base,
    ...(metric === "up" || metric === "all"
      ? [
          {
            title: "当日涨跌幅",
            dataIndex: "change",
            render: (v?: number) => (v === undefined ? "—" : format(v) + "%"),
          },
        ]
      : []),
    ...(metric === "high" || metric === "all" ? price : []),
    ...(metric === "all" ? [{ title: "一年新高", dataIndex: "newHigh", render: yes }] : []),
    ...(metric === "csi" || metric === "all"
      ? [
          { title: "MA50", dataIndex: "ma50", render: format },
          { title: "高于均线", dataIndex: "above", render: yes },
        ]
      : []),
    { title: "有效日线（最多 250 根）", dataIndex: "bars" },
    ...(metric === "excluded" || metric === "all" || metric === "csi"
      ? [{ title: "说明", dataIndex: "reason", render: (v?: string) => v || "—" }]
      : []),
  ];
  return (
    <Drawer title={day.date + " · " + labels[metric]} open onClose={close} width="min(1080px, 100vw)">
      <Input.Search
        aria-label="搜索明细"
        placeholder="按股票代码或名称查询"
        allowClear
        onSearch={value => {
          setSearch(value.trim());
          setPage(1);
        }}
      />
      {list.isError ? (
        <Alert type="error" message="明细读取失败" action={<Button onClick={() => void list.refetch()}>重试</Button>} />
      ) : (
        <Table<Row>
          size="small"
          rowKey="symbol"
          columns={columns}
          dataSource={list.data?.rows ?? []}
          loading={list.isPending}
          pagination={false}
          scroll={{ x: "max-content" }}
          locale={{ emptyText: "没有匹配的股票" }}
        />
      )}
      {!!list.data && (
        <Pagination
          current={page}
          pageSize={50}
          total={list.data.total}
          showSizeChanger={false}
          showTotal={n => "共 " + n + " 家"}
          onChange={setPage}
        />
      )}
    </Drawer>
  );
}

const apiLabels: Record<string, string> = {
  "baostock.com": "BaoStock", "quantapi.51ifind.com": "同花顺 iFinD", "help.tdx.com.cn": "通达信",
  "quantapi.eastmoney.com": "东方财富 Choice", "tushare.pro": "Tushare Pro",
  "stock.qq.com": "腾讯财经", "finance.sina.com.cn": "新浪财经",
};

function reviewSourceLabel(source: string) {
  const host = URL.canParse(source) ? new URL(source).hostname : "";
  return apiLabels[host] ?? (host === "api.zizizaizai.com" ? "自在量化" : host || source);
}

function isApiReview(source: string) {
  return URL.canParse(source) && new URL(source).hostname in apiLabels;
}

function Daily({ day, delta, selectDate }: { day: Day; delta: number | null; selectDate: (date: string) => void }) {
  const [metric, setMetric] = useState<Metric>();
  const [params, setParams] = useSearchParams();
  const requested = params.get("section");
  const sections = ["history", "industry", "style", "patterns"];
  const hasPatternFilter = ["stage", "pattern", "search", "big", "near", "width", "doji", "shrink"].some(key => params.has(key));
  const section = requested && sections.includes(requested) ? requested : hasPatternFilter ? "patterns" : "history";
  function selectSection(key: string) {
    const next = new URLSearchParams(params);
    next.set("section", key);
    setParams(next, {replace: true});
  }
  const {focus, others, change, ratio, gaps} = reviewGuidance(day, delta);
  const focusMetric = focus.metric;
  const figures = [
    { title: "上涨家数", value: day.up + " 家", effect: day.traded + " 家有当日成交日线", metric: "up" },
    {
      title: "一年新高家数",
      value: day.highs + " 家",
      effect: change,
      metric: "high",
    },
    {
      title: "沪深 300 的 50 日占比",
      value: ratio === null ? "—" : ratio.toFixed(2) + "%",
      effect: day.above + " 家高于均线 / " + day.eligible + " 家有效成份",
      metric: "csi",
    },
  ] as const;
  return (
    <>
      <section className={styles.brief} aria-label="当日复盘重点">
        <div className={styles.briefHeading}>{day.date} · 收盘观察</div>
        <h2>{focus.title}</h2>
        <p>{focus.reason}</p>
        <div className={styles.nextStep}><strong>复盘建议</strong><span>{focus.next}</span>
          {focusMetric ? <button onClick={() => setMetric(focusMetric)}>{labels[focusMetric]} →</button>
            : <button onClick={() => selectSection("industry")}>查看行业表 →</button>}
        </div>
        {others.map(rule => <div key={rule.title} className={styles.secondary}>
          <strong>{rule.title}</strong>　{rule.reason}{rule.next}
          <button onClick={() => setMetric(rule.metric)}>查看名单 →</button>
        </div>)}
        {gaps && <p>数据缺口 · {gaps}</p>}
      </section>
      <div className={styles.figures}>
        {figures.map(f => (
          <Card key={f.metric}>
            <h2>{f.title}</h2>
            <strong>{f.value}</strong>
            <p>{f.effect}</p>
            <Button type="link" onClick={() => setMetric(f.metric)}>
              {labels[f.metric]} →
            </Button>
          </Card>
        ))}
      </div>
      <details className={styles.method}>
        <summary>按原文判断操作 · 从观望到买入或退出</summary>
        <p><strong>空仓时先看环境。</strong>一年新高不足 20 家，作者通常暂停强势股交易；重新超过 20 家，恢复找机会。再对照近期新高指数与平均股价的强弱，以及近期新高指数相对 EMA5、EMA13 的位置。</p>
        <p><strong>有方向才选股。</strong>跟踪几天里反复进入容量、强度前排的行业。优先复核其中成交活跃、前段涨幅突出、在题材中有辨识度的股票。</p>
        <p><strong>候选股等分歧。</strong>作者关注强股上涨途中的十字星。结合前期大阳、横盘或回踩的位置，确认支撑和板块呼应，并能事先写出失效条件，才进入买入评估。</p>
        <p><strong>持仓按计划处理。</strong>作者在板块与个股重新一致走强时考虑兑现；触及预设止损时退出。文章没有为所有个股给出统一卖价。</p>
        <p>页面的“当日触发”采用项目筛选口径，即下一交易日最高价突破候选日最高价。题材逻辑、承接质量和个人风险预算仍需自行核对。</p>
      </details>
      <Tabs activeKey={section} onChange={selectSection} items={[
        {key:"history",label:"逐日对照",children:<ReviewTimeline date={day.date} select={selectDate}/>},
        {key:"industry",label:"行业容量与强度",children:<SectorCard day={day}/>},
        {key:"style",label:"风格指数",children:<StyleCard day={day}/>},
        {key:"patterns",label:"个股形态",children:<PatternCard day={day}/>},
      ]} />
      <details className={styles.method}><summary>统计口径、排除明细与来源</summary>
        <p>
          沪深 A 股（含 ST）。当日股票池 {day.total} 家，{day.total - day.traded}{" "}
          家无当日有效成交日线。上涨按收盘日涨跌幅大于零计。
        </p>
        <p>
          一年新高取当日最高价严格超过此前 249 根有效日线最高价，历史价格按当日复权基准比较。有效日线不足的有{" "}
          {day.traded - day.highEligible} 家。
        </p>
        <p>
          沪深 300 采用当日成份 {day.csi} 家。MA50 包含当日收盘，排除 {day.csi - day.eligible} 家无有效样本的成份股。
        </p>
        <p>
          {delta === null
            ? "前一交易日 " + day.previous + " 暂无同口径归档可比较。"
            : "较上一交易日 " + day.previous + "，新高家数变化 " + (delta > 0 ? "+" : "") + delta + " 家。"}
          {delta !== null && delta > 0 && day.up < 1000 && " 当日上涨不足 1,000 家且新高增加，符合普跌背离观察条件。"}
        </p>
        <div className={styles.actions}>
          <Button onClick={() => setMetric("all")}>全部样本</Button>
          <Button onClick={() => setMetric("excluded")}>查看排除明细</Button>
        </div>

      <footer>
        当日行情{" "}
        <a href={day.source} target="_blank" rel="noreferrer">
          {reviewSourceLabel(day.source)}
        </a>{" "}
        {isApiReview(day.source) ? " · 交易日、历史个股行情及复权因子来自 BaoStock" : <>
          · 历史行情与交易日核对{" "}
          <a href={"https://github.com/chenditc/investment_data/releases/tag/" + day.release}
            target="_blank" rel="noreferrer">investment_data {day.release}</a>
        </>}
        {(day.source.includes("/daily_kline/date/") || day.source.includes("/v3/market/kline/day?")) && (
          <>
            {" "}
            · 当日成份核对{" "}
            {["159919", "159925"].map(fund => (
              <a
                key={fund}
                href={`https://reportdocs.static.szse.cn/files/text/ETFDown/pcf_${fund}_${day.date.replaceAll("-", "")}.xml`}
                target="_blank"
                rel="noreferrer"
              >
                {" "}
                {fund}{" "}
              </a>
            ))}
          </>
        )}
        <br />
        规则 {day.rule} · 批次 {day.batch.slice(0, 12)} · 预热起点 {day.windowStart}
      </footer>
      </details>
      {metric && <Members day={day} metric={metric} close={() => setMetric(undefined)} />}
    </>
  );
}

export default function MarketReview() {
  const [params, setParams] = useSearchParams();
  const dates = useQuery({
    queryKey: ["review-dates"],
    queryFn: async ({ signal }) =>
      (await RequestGet("/api/v1/review/dates", { schema: z.array(z.string()), signal })).data,
  });
  const date = params.get("date") ?? dates.data?.[0];
  function selectDate(value: string) {
    const next = new URLSearchParams(params);
    next.set("date", value);
    setParams(next);
  }
  const daily = useQuery({
    queryKey: ["review-day", date],
    enabled: !!date,
    queryFn: async ({ signal }) =>
      (await RequestGet("/api/v1/review/daily", { params: { date }, schema: reportSchema, signal })).data,
  });
  return (
    <main className={styles.page}>
      <header>
        <div>
          <h1>复盘</h1>
          <p>沪深 A 股 · 收盘归档</p>
        </div>
        <label>
          复盘交易日{" "}
          <Select
            aria-label="复盘交易日"
            showSearch
            value={date}
            placeholder="选择归档日"
            loading={dates.isPending}
            options={dates.data?.map(value => ({ label: value, value }))}
            onChange={selectDate}
          />
        </label>
        <ReviewUpdate selectDate={selectDate} requestedDate={params.get("date") ?? undefined}/>
      </header>
      {dates.isError || daily.isError ? (
        <Alert
          type="error"
          message="无法读取复盘归档"
          action={
            <Button
              onClick={() => {
                void dates.refetch();
                if (date) void daily.refetch();
              }}
            >
              重试
            </Button>
          }
        />
      ) : dates.isPending || (date && daily.isPending) ? (
        <Skeleton active />
      ) : daily.data ? (
        <Daily key={daily.data.day.batch} day={daily.data.day} delta={daily.data.highDelta} selectDate={selectDate} />
      ) : (
        <><Empty description={date ? date + " 暂无已发布复盘" : "暂无已发布复盘"} />{date && <ReviewTimeline date={date} select={selectDate} />}</>
      )}
    </main>
  );
}

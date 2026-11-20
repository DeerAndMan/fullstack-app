import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import { Alert, DatePicker } from "antd";
import dayjs from "dayjs";
import { displayMarketDate, hotspotDatePicker } from "./calendar";
import { Change, DataTable, FeedPanel } from "./panels";
import { localDate, useMarketFeed } from "./data";
import { SourcePicker, useMarketSource } from "./source";
import type { MarketRow } from "./data";
import styles from "./market-history.module.css";

const bands = [
  { name: "冰点", color: "#238668" },
  { name: "一般", color: "#9e972d" },
  { name: "良好", color: "#d48c26" },
  { name: "火热", color: "#c65b62" },
];
const temperature = (n: number) => bands[Math.min(3, Math.floor(n / 25))].name;
const money = (row: MarketRow) => row.amount == null ? "—" : (row.amount / 1e8).toLocaleString("zh-CN", { maximumFractionDigits: 2 });
const upColor = "#c74e50",
  downColor = "#228568",
  volumeColor = "#547ab0",
  scoreColor = "#c08020";
const scoreSeries = [
  { name: "开盘啦情绪指标", color: scoreColor },
  { name: "A股每日复盘", color: "#397c94" },
  { name: "知市综合情绪（第一版）", color: "#835eaa" },
];
function HistoryTrends({ rows, amountOnly }: { rows: MarketRow[]; amountOnly: boolean }) {
  const [focus, setFocus] = useState("");
  const selected = rows.find(r => r.date === focus) ?? rows[rows.length - 1];
  const archiveDates = new Set(rows.map(row => row.date));
  const left = 56,
    right = 908,
    width = right - left;
  const slot = width / rows.length;
  const x = (i: number) => left + (i + 0.5) * slot;
  const scoreY = (r: MarketRow) => r.sentiment == null ? undefined : 146 - r.sentiment * 1.2;
  const maxAmount = Math.max(1, ...rows.map(r => (r.amount ?? 0) / 1e8));
  const maxCount = Math.max(1, ...rows.flatMap(r => [r.advancers ?? 0, r.decliners ?? 0]));
  const volumeY = (r: MarketRow) => r.amount == null ? undefined : 295 - (r.amount / 1e8 / maxAmount) * 95;
  function line(y: (row: MarketRow) => number | undefined, source?: string) {
    let connected = false;
    return rows.map((row, i) => {
      const value = y(row);
      if (value === undefined || (source && row.detail !== source)) {
        connected = false;
        return "";
      }
      const command = connected ? "L" : "M";
      connected = true;
      return command + x(i) + "," + value;
    }).join(" ");
  }
  const interval = Math.max(1, Math.ceil(rows.length / 7));
  const amountLabel = amountOnly ? "上证综指与深证综指成交额合计" : "沪深成交额";
  return (
    <>
      <div className={styles.readout}>
        <div className={styles.archivePicker}>
          <label htmlFor="market-archive-date">归档日期</label>
          <DatePicker
            {...hotspotDatePicker}
            id="market-archive-date"
            allowClear={false}
            showToday={false}
            value={dayjs(selected.date)}
            minDate={dayjs(rows[0].date)}
            maxDate={dayjs(rows[rows.length - 1].date)}
            disabledDate={(date, info) => info.type === "date" && !archiveDates.has(date.format("YYYY-MM-DD"))}
            onChange={date => setFocus(date.format("YYYY-MM-DD"))}
          />
        </div>
        {!amountOnly && <div>
          {selected.detail ?? "情绪指标"}{" "}
          <strong>
            {selected.sentiment == null ? "—" : `${selected.sentiment} · ${temperature(selected.sentiment)}`}
          </strong>
        </div>}
        <div>
          {amountLabel} <strong>{money(selected)} 亿</strong>
        </div>
        {!amountOnly && <>
          <div>
            上涨 <strong style={{ color: upColor }}>{selected.advancers ?? "—"} 家</strong>
          </div>
          <div>
            下跌 <strong style={{ color: downColor }}>{selected.decliners ?? "—"} 家</strong>
          </div>
          {selected.quoteSource && selected.quoteSource !== "开盘啦" && <span>成交额与家数来自{selected.quoteSource}</span>}
        </>}
      </div>
      <div className={styles.scroll}>
        <svg viewBox={amountOnly ? "0 163 980 175" : "0 0 980 385"} role="img"
          aria-label={amountOnly ? amountLabel : "历史市场情绪、沪深成交额与上涨下跌家数。逐日数值见下方明细。"}>
          {!amountOnly && <>
            <text x={left} y={15} className={styles.title}>
              市场情绪（分）
            </text>
            {bands.map((b, i) => (
              <g key={b.name}>
                <rect x={left} y={116 - i * 30} width={width} height={30} fill={b.color} fillOpacity=".10" />
                <text x={right + 8} y={136 - i * 30}>
                  {b.name}
                </text>
              </g>
            ))}
            {[0, 25, 50, 75, 100].map(n => (
              <text x={left - 8} textAnchor="end" y={150 - n * 1.2} key={n}>
                {n}
              </text>
            ))}
            {scoreSeries.map(series => (
              <path key={series.name} d={line(scoreY, series.name)} fill="none" stroke={series.color} strokeWidth="2.4" />
            ))}
          </>}
          <text x={left} y={184} className={styles.title}>
            {amountLabel}（亿元）
          </text>
          {!amountOnly && <text x={right} y={184} textAnchor="end">
            上涨 / 下跌（家）
          </text>}
          <line x1={left} x2={right} y1="295" y2="295" stroke="currentColor" opacity=".4" />
          <text x={left - 6} y="204" textAnchor="end">
            {Math.round(maxAmount)}
          </text>
          <text x={left - 6} y="299" textAnchor="end">
            0
          </text>
          {!amountOnly && <>
            <text x={right + 7} y="240">
              {maxCount}
            </text>
            <text x={right + 7} y="299">
              0
            </text>
            <text x={right + 7} y="357">
              {maxCount}
            </text>
          </>}
          <path d={line(volumeY)} fill="none" stroke={volumeColor} strokeWidth="2" strokeDasharray="5 3" />
          {rows.map((r, i) => {
            const hi = ((r.advancers ?? 0) / maxCount) * 58,
              lo = ((r.decliners ?? 0) / maxCount) * 58;
            const w = Math.min(19, slot * 0.47);
            return (
              <g key={r.code} onPointerEnter={() => setFocus(r.date!)} onClick={() => setFocus(r.date!)}>
                {selected === r && (
                  <line
                    x1={x(i)}
                    x2={x(i)}
                    y1="25"
                    y2="354"
                    stroke="currentColor"
                    opacity=".26"
                    strokeDasharray="3 4"
                  />
                )}
                {r.sentiment != null && <circle cx={x(i)} cy={scoreY(r)} r={selected === r ? 4.5 : 3} fill={scoreSeries.find(series => series.name === r.detail)?.color ?? scoreColor} />}
                {r.amount != null && <circle cx={x(i)} cy={volumeY(r)} r="3" fill={volumeColor} />}
                <rect x={x(i) - w / 2} y={295 - hi} width={w} height={hi} fill={upColor} />
                <rect x={x(i) - w / 2} y="295" width={w} height={lo} fill={downColor} />
                {(i === rows.length - 1 || (i % interval === 0 && rows.length - 1 - i >= Math.ceil(interval / 2))) && (
                  <text x={x(i)} y={amountOnly ? 320 : 376} textAnchor="middle">
                    {displayMarketDate(r.date, false)}
                  </text>
                )}
                <rect x={x(i) - slot / 2} y="25" width={slot} height="330" fill="transparent">
                  <title>
                    {displayMarketDate(r.date)} · 成交 {money(r)} 亿
                    {!amountOnly && ` · 情绪 ${r.sentiment ?? "—"} · 涨 ${r.advancers ?? "—"} / 跌 ${r.decliners ?? "—"}`}
                  </title>
                </rect>
              </g>
            );
          })}
        </svg>
      </div>
      {!amountOnly && <p className={styles.note}>
        {scoreSeries.filter(series => rows.some(r => r.detail === series.name)).map(series => (
          <span key={series.name} style={{ color: series.color }}>● {series.name}　</span>
        ))}
        蓝线为沪深成交额，红柱为上涨家数，绿柱为下跌家数。
      </p>}
      <details>
        <summary>{amountOnly ? "逐日成交额" : "逐日数据与评分口径"}</summary>
        {!amountOnly && <p>
          主源为开盘啦。情绪分采用开盘啦公布的历史情绪指标，沪深成交额和涨跌家数取同日收盘统计。
          同花顺上证指数日线用于核对交易日。主源缺项时优先使用已保存的开盘啦收盘数据，其次使用同日的备用来源。
          知市和 A股每日复盘各自计算情绪分，按来源分段绘制。四档分界为 25、50、75 分。

        </p>}
        <DataTable
          rows={rows
            .slice()
            .reverse()
            .map(r => ({ ...r, amount: r.amount == null ? undefined : Number((r.amount / 1e8).toFixed(2)) }))}
          label="市场历史冷热"
          columns={[
            { title: "归档日期", field: "date", render: row => displayMarketDate(row.date) },
            ...(!amountOnly ? [
              { title: "情绪分", field: "sentiment" },
              {
                title: "情绪来源", field: "detail",
                render: (r: MarketRow) => r.scoreUrl ? <a href={r.scoreUrl} target="_blank" rel="noreferrer">{r.detail}</a> : "—",
              },
            ] as const : []),
            { title: `${amountLabel}（亿元）`, field: "amount" },
            ...(!amountOnly ? [
              { title: "上涨家数", field: "advancers" },
              { title: "下跌家数", field: "decliners" },
            ] as const : []),
            {
              title: amountOnly ? "成交额来源" : "成交额与家数来源",
              field: "url",
              render: r => r.url ? (
                <a href={r.url} target="_blank" rel="noreferrer">
                  {r.quoteSource}
                </a>
              ) : "—",
            },
          ]}
        />
      </details>
    </>
  );
}

const indexSeries = [
  { name: "上证综指", field: "shanghaiChange", color: "#397c94" },
  { name: "深证综指", field: "shenzhenChange", color: "#835eaa" },
] as const;

function IndexChanges({ rows }: { rows: MarketRow[] }) {
  const maxChange = Math.max(1, ...rows.flatMap(row => indexSeries.map(series => Math.abs(row[series.field] ?? 0))));
  const x = (index: number) => rows.length < 2 ? 480 : 55 + 855 * index / (rows.length - 1);
  const y = (change: number) => 104 - 78 * change / maxChange;
  const tickIndices = new Set(Array.from({ length: 5 }, (_, i) => Math.round(i * (rows.length - 1) / 4)));
  return <>
    <div className={styles.scroll}>
      <svg viewBox="0 0 960 228" role="img" aria-label="上证综指与深证综指日涨跌幅，逐日数值见下方明细。">
        {[maxChange, 0, -maxChange].map(change => <g key={change}>
          <line x1="55" x2="910" y1={y(change)} y2={y(change)} stroke="currentColor" opacity=".22" />
          <text x="47" y={y(change) + 4} textAnchor="end">{change.toFixed(1)}%</text>
        </g>)}
        {indexSeries.map(series => <polyline key={series.field}
          points={rows.map((row, index) => `${x(index)},${y(row[series.field] ?? 0)}`).join(" ")}
          fill="none" stroke={series.color} strokeWidth="2" />)}
        {rows.length === 1 && indexSeries.map(series => <circle key={series.field}
          cx={x(0)} cy={y(rows[0][series.field] ?? 0)} r="4" fill={series.color} />)}
        {[...tickIndices].map(index => <text key={index} x={x(index)} y="209" textAnchor="middle">
          {displayMarketDate(rows[index].date, false)}
        </text>)}
      </svg>
    </div>
    <p className={styles.note}>{indexSeries.map(series => <span key={series.field} style={{ color: series.color }}>● {series.name}　</span>)}</p>
    <details>
      <summary>逐日指数涨跌幅</summary>
      <DataTable label="逐日指数涨跌幅" rows={rows.slice().reverse()} columns={[
        { title: "交易日", field: "date" },
        ...indexSeries.map(series => ({ title: series.name, field: series.field, render: (row: MarketRow) => <Change value={row[series.field]} /> })),
      ]} />
    </details>
  </>;
}

const historySources = [
  { value: "market-history", label: "开盘啦及备用来源" },
  { value: "market-history-sina", label: "新浪财经" },
  { value: "market-history-eastmoney", label: "东方财富" },
  { value: "market-history-ths", label: "同花顺" },
] as const;

export function MarketHistory() {
  const [params, setParams] = useSearchParams();
  const from = params.get("from");
  const to = params.get("to");
  const period = from && to ? from + ".." + to : "";
  const today = dayjs(localDate());
  const [feed, setFeed] = useMarketSource("historySource", historySources);
  const query = useMarketFeed(feed, period ? { date: period } : undefined);
  const indexChanges = feed === "market-history-sina";
  const amountOnly = feed === "market-history-ths" || feed === "market-history-eastmoney";
  return (
    <FeedPanel
      title="市场历史冷热"
      feed={feed}
      query={query}
      extra={<SourcePicker label="历史冷热来源" value={feed} options={historySources} onChange={setFeed} />}
      note={indexChanges
        ? "按新浪财经指数日线收盘价计算上证综指和深证综指的日涨跌幅。"
        : amountOnly ? "成交额取上证综指和深证综指日线之和。情绪分与涨跌家数可在开盘啦来源查看。" : undefined}
      controls={
        <div className={styles.filters}>
          <span>查询区间</span>
          <DatePicker.RangePicker
            {...hotspotDatePicker}
            separator={<span aria-label="至">→</span>}
            placeholder={["开始日期", "结束日期"]}
            value={from && to ? [dayjs(from), dayjs(to)] : null}
            disabledDate={d => d.isAfter(today, "day")}
            presets={[
              { label: "本月", value: [today.startOf("month"), today] },
              { label: "上月", value: [today.subtract(1, "month").startOf("month"), today.subtract(1, "month").endOf("month")] },
            ]}
            onChange={range => {
              const next = new URLSearchParams(params);
              if (range?.[0] && range[1]) {
                next.set("from", range[0].format("YYYY-MM-DD"));
                next.set("to", range[1].format("YYYY-MM-DD"));
              } else {
                next.delete("from");
                next.delete("to");
              }
              setParams(next);
            }}
          />
          {!period && <span>最近 20 个归档日</span>}
          {query.data?.rows.length ? (
            <span>
              {query.data.rows[0].date} 至 {query.data.rows.at(-1)?.date} · {query.data.rows.length} 个归档日
            </span>
          ) : null}
          {query.data?.message && <Alert showIcon type="warning" message={query.data.message} />}
        </div>
      }
    >
      {rows => indexChanges ? <IndexChanges rows={rows} /> : (
        <HistoryTrends key={`${feed}:${period}`} rows={rows} amountOnly={amountOnly} />
      )}
    </FeedPanel>
  );
}

import { useEffect, useMemo, useRef, useState } from "react";
import { Empty, Select, Segmented } from "antd";
import dayjs from "dayjs";

import { useMarketFeed } from "./data";
import type { MarketRow } from "./data";
import { FeedPanel } from "./panels";
import styles from "./stock-valuation.module.css";

const metrics = {
  peTtm: { label: "市盈率", basis: "滚动十二个月" },
  pbMrq: { label: "市净率", basis: "最近一期财报" },
  psTtm: { label: "市销率", basis: "滚动十二个月" },
  pcfOcfTtm: { label: "市现率", basis: "近十二个月经营现金流" },
} as const;
type Metric = keyof typeof metrics;
const ranges = [1, 3, 5, 10];
const format = new Intl.NumberFormat("zh-CN", { maximumFractionDigits: 2 });
const multiple = (value?: number) => value === undefined ? "—" : `${format.format(value)} 倍`;
const top = 16;
const bottom = 190;
const chartHeight = 224;

function percentile(sorted: number[], fraction: number) {
  const index = (sorted.length - 1) * fraction;
  const lower = Math.floor(index);
  return sorted[lower] + (sorted[Math.ceil(index)] - sorted[lower]) * (index - lower);
}

function ValuationChart({ dates, values, label, low, high }: {
  dates: string[];
  values: (number | undefined)[];
  label: string;
  low: number;
  high: number;
}) {
  const svg = useRef<SVGSVGElement>(null);
  const [width, setWidth] = useState(640);
  const [active, setActive] = useState<number>();
  useEffect(() => {
    const element = svg.current;
    if (!element) return;
    const observer = new ResizeObserver(([entry]) => setWidth(Math.max(180, entry.contentRect.width)));
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  const maximum = Math.max(...values.map(value => value ?? 0));
  const magnitude = 10 ** Math.floor(Math.log10(maximum / 4));
  const step = Math.ceil(maximum / 4 / magnitude) * magnitude;
  const left = Math.max(43, format.format(step * 4).length * 7 + 13);
  const right = width - 14;
  const last = dates.length - 1;
  const x = (index: number) => last ? left + index / last * (right - left) : (left + right) / 2;
  const y = (value: number) => bottom - value / (step * 4) * (bottom - top);
  const index = Math.min(active ?? last, last);
  const value = values[index];

  let previousDefined = false;
  const line = values.map((point, i) => {
    const command = previousDefined ? "L" : "M";
    previousDefined = point !== undefined;
    return point === undefined ? "" : `${command}${x(i).toFixed(2)},${y(point).toFixed(2)}`;
  }).join(" ");

  function selectAt(clientX: number) {
    const rect = svg.current?.getBoundingClientRect();
    if (!rect) return;
    const position = (clientX - rect.left - left) / (right - left);
    setActive(Math.min(last, Math.max(0, Math.round(position * last))));
  }

  return <>
    <div className={styles.readout}>
      <span><i className={styles.trend} />{label} <strong>{multiple(value)}</strong></span>
      <time dateTime={dates[index]}>{dates[index]}</time>
    </div>
    <svg
      ref={svg}
      className={styles.chart}
      width="100%"
      height={chartHeight}
      role="slider"
      tabIndex={0}
      aria-label={`${label}走势，左右方向键查看交易日`}
      aria-valuemin={0}
      aria-valuemax={last}
      aria-valuenow={index}
      aria-valuetext={`${dates[index]} ${multiple(value)}`}
      onPointerMove={event => event.pointerType === "mouse" && selectAt(event.clientX)}
      onPointerDown={event => selectAt(event.clientX)}
      onKeyDown={event => {
        if (event.key === "ArrowLeft" || event.key === "ArrowRight") {
          event.preventDefault();
          setActive(Math.min(last, Math.max(0, index + (event.key === "ArrowLeft" ? -1 : 1))));
        }
      }}
    >
      <title>{label}历史走势</title>
      {[0, 1, 2, 3, 4].map(tick => (
        <g key={tick}>
          <line x1={left} x2={right} y1={y(tick * step)} y2={y(tick * step)} className={styles.grid} />
          <text x={left - 8} y={y(tick * step)} textAnchor="end" dominantBaseline="middle">{format.format(tick * step)}</text>
        </g>
      ))}
      <rect x={left} y={y(high)} width={right - left} height={y(low) - y(high)} className={styles.band} />
      <line x1={left} x2={right} y1={y(low)} y2={y(low)} className={styles.lowLine} />
      <line x1={left} x2={right} y1={y(high)} y2={y(high)} className={styles.highLine} />
      <path d={line} className={styles.series} />
      {last > 0 && <text x={left} y={bottom + 23}>{dates[0]}</text>}
      <text x={right} y={bottom + 23} textAnchor="end">{dates[last]}</text>
      <line x1={x(index)} x2={x(index)} y1={top} y2={bottom} className={styles.cursor} />
      {value !== undefined && <circle cx={x(index)} cy={y(value)} r="3.5" className={styles.point} />}
    </svg>
  </>;
}

function ValuationHistory({ rows, metric, years }: { rows: MarketRow[]; metric: Metric; years: number }) {
  const latest = rows[rows.length - 1];
  const start = dayjs(latest.date).subtract(years, "year").format("YYYY-MM-DD");
  const { dates, values, sorted } = useMemo(() => {
    const window = rows.filter(row => row.date && row.date >= start);
    const dates = window.map(row => row.date!);
    const values = window.map(row => {
      const value = row[metric];
      return value !== undefined && value > 0 ? value : undefined;
    });
    return { dates, values, sorted: values.filter(value => value !== undefined).toSorted((a, b) => a - b) };
  }, [metric, rows, start]);

  const current = latest[metric];
  const rank = current !== undefined && current > 0
    ? sorted.filter(value => value <= current).length / sorted.length * 100
    : undefined;
  const firstValidDate = dates[values.findIndex(value => value !== undefined)];
  const lastValidDate = dates[values.findLastIndex(value => value !== undefined)];
  const { label, basis } = metrics[metric];

  return <>
    <div className={styles.summary}>
      <div>
        <span>{label} · {basis}</span>
        <strong>{multiple(current)}</strong>
      </div>
      <div>
        <span>历史分位</span>
        <strong>{rank === undefined ? "—" : `${format.format(rank)}%`}</strong>
      </div>
      <div>
        <span>估值日期</span>
        <time dateTime={latest.date}>{latest.date}</time>
      </div>
    </div>
    {sorted.length ? <>
      <ValuationChart key={`${metric}:${years}`} dates={dates} values={values} label={label}
        low={percentile(sorted, 0.3)} high={percentile(sorted, 0.7)} />
      <div className={styles.legend}>
        <span><i className={styles.low} />30% 分位 <strong>{multiple(percentile(sorted, 0.3))}</strong></span>
        <span><i className={styles.high} />70% 分位 <strong>{multiple(percentile(sorted, 0.7))}</strong></span>
        <span className={styles.coverage}>数据覆盖 {firstValidDate} 至 {lastValidDate} · {sorted.length.toLocaleString("zh-CN")} 个有效交易日</span>
      </div>
    </> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="该区间暂无可比较的正值估值" />}
    <details className={styles.method}>
      <summary>估值和分位口径</summary>
      <p>历史分位是所选区间中，小于或等于最新估值的有效交易日占比。走势和分位使用东方财富提供的正值，缺失或非正值处留空。最新估值按数据源披露值展示。</p>
    </details>
  </>;
}

export function StockValuation({ symbol }: { symbol: string }) {
  const [metric, setMetric] = useState<Metric>("peTtm");
  const [years, setYears] = useState(10);
  const query = useMarketFeed("stock-valuation", { symbol });

  return <section aria-label="估值分析">
    <FeedPanel title="估值分析" feed="stock-valuation" query={query} controls={
      <div className={styles.toolbar}>
        <Segmented size="small" aria-label="估值指标" value={metric} onChange={setMetric}
          options={(Object.keys(metrics) as Metric[]).map(value => ({ value, label: metrics[value].label }))} />
        <Select size="small" aria-label="估值区间" value={years} onChange={setYears}
          options={ranges.map(value => ({ value, label: `近 ${value} 年` }))} />
      </div>
    }>
      {rows => <ValuationHistory rows={rows} metric={metric} years={years} />}
    </FeedPanel>
  </section>;
}

import { Alert, Button, Card, Empty, Segmented, Skeleton, theme } from "antd";
import { ArrowLeftOutlined } from "@ant-design/icons";
import { Link, useLocation, useParams, useSearchParams } from "react-router-dom";

import ROUTER_PATH from "@/router/section/router-path";
import { ConceptReason } from "./concept-reason";
import { isStockSymbol, moneyText, priceText, updateTime, useMarketFeed } from "./data";
import { Change, withTokenColors } from "./panels";
import { StockChart, volumeText } from "./stock-chart";
import { StockValuation } from "./stock-valuation";
import styles from "./hotspots.module.css";

import type { MarketQuery } from "./data";

const periods = [
  { value: "stock-minute", label: "分时" },
  { value: "stock-five-day", label: "五日" },
  { value: "stock-daily", label: "日 K" },
  { value: "stock-weekly", label: "周 K" },
  { value: "stock-monthly", label: "月 K" },
] as const;

function QuoteState({ query, label }: { query: MarketQuery; label: string }) {
  if (query.isPending) return <Skeleton active title={false} paragraph={{ rows: 2 }} />;
  if (query.isError || query.data?.status === "unavailable") {
    return <Alert type="warning" showIcon message={`暂时无法获取${label}`}
      action={<Button size="small" loading={query.isFetching} onClick={() => void query.refetch()}>重试</Button>} />;
  }
  if (query.data?.status === "stale") {
    return <Alert type="warning" showIcon message={`${label}更新失败，显示最近一次可用数据。`} />;
  }
  return null;
}

function StockDetail({ symbol, stockName, topic, topicName }: {
  symbol: string;
  stockName?: string;
  topic?: string;
  topicName?: string;
}) {
  const [params, setParams] = useSearchParams();
  const { state } = useLocation();
  const period = periods.find(item => item.value === params.get("period")) ?? periods[0];
  const quote = useMarketFeed("stock-quote", { symbol });
  const chart = useMarketFeed(period.value, { symbol });
  const snapshot = quote.data?.rows[0];
  const previous = snapshot?.previous ? Number(snapshot.previous) : NaN;
  const rows = chart.data?.rows ?? [];
  const intraday = period.value === "stock-minute" || period.value === "stock-five-day";
  const lastDate = rows.at(-1)?.date?.slice(0, 10);
  const reference = period.value === "stock-minute" && lastDate && lastDate === snapshot?.date?.slice(0, 10) &&
    Number.isFinite(previous) && previous > 0 ? previous : undefined;

  return <>
    <Card className={styles.stockQuote}>
      <header className={styles.stockHeader}>
        <div>
          <h1>{snapshot?.name || stockName || symbol.slice(2)}</h1>
          <span className={styles.muted}>{symbol.toUpperCase()}</span>
        </div>
        {snapshot && <div className={styles.stockHeadline}>
          <strong>{priceText(snapshot.price)}</strong>
          <Change value={snapshot.change} />
        </div>}
        {snapshot?.date && <span className={styles.muted}>报价 {snapshot.date}（北京时间）</span>}
      </header>
      <QuoteState query={quote} label="个股报价" />
      {snapshot && <dl className={styles.stockMetrics}>
        <div><dt>今开</dt><dd>{priceText(snapshot.open)}</dd></div>
        <div><dt>昨收</dt><dd>{priceText(Number.isFinite(previous) ? previous : undefined)}</dd></div>
        <div><dt>最高</dt><dd>{priceText(snapshot.high)}</dd></div>
        <div><dt>最低</dt><dd>{priceText(snapshot.low)}</dd></div>
        <div><dt>成交量</dt><dd>{volumeText(snapshot.volume)}</dd></div>
        <div><dt>成交额</dt><dd>{moneyText(snapshot.amount)}</dd></div>
      </dl>}
    </Card>
    <Card className={styles.stockChartCard}>
      {topic && <section className={styles.stockConcept} aria-labelledby="stock-concept-title">
        <h2 id="stock-concept-title">{topicName ? `${topicName} · 关联原因` : "关联原因"}</h2>
        <ConceptReason concept={topic} symbol={symbol} />
      </section>}
      <div className={styles.stockChartToolbar}>
        <Segmented aria-label="行情周期" value={period.value} options={periods.map(item => ({ ...item }))}
          onChange={value => setParams(previous => {
            const next = new URLSearchParams(previous);
            if (value === periods[0].value) next.delete("period");
            else next.set("period", value);
            return next;
          }, { replace: true, state })} />
        {chart.data?.adjustment && <span className={styles.muted}>{chart.data.adjustment === "qfq" ? "前复权" : "不复权"}</span>}
      </div>
      <QuoteState query={chart} label={`${period.label}行情`} />
      {rows.length > 0 && <StockChart key={period.value} rows={rows} label={period.label} intraday={intraday} previousClose={reference} />}
      {chart.data?.status === "ready" && rows.length === 0 &&
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="该周期暂无行情数据" />}
      <div className={styles.source}>
        <a href={`https://gu.qq.com/${symbol}`} target="_blank" rel="noreferrer">腾讯财经</a>
        {chart.data?.status && chart.data.status !== "unavailable" &&
          <span>获取于 {updateTime(chart.data.fetchedAt)}（北京时间）</span>}
      </div>
    </Card>
    <StockValuation symbol={symbol} />
  </>;
}

export default function StockDetailPage() {
  const { token } = theme.useToken();
  const { symbol } = useParams();
  const [params] = useSearchParams();
  const location = useLocation();
  const { stockName } = (location.state ?? {}) as { stockName?: string };
  const back = new URLSearchParams(params);
  back.delete("period");
  const topicParam = params.get("topic");
  const topic = topicParam && /^gn_[a-z0-9_]{1,40}$/.test(topicParam) ? topicParam : undefined;
  const catalog = useMarketFeed("concept", undefined, Boolean(topic && isStockSymbol(symbol)));
  const topicName = catalog.data?.rows.find(row => row.code === topic)?.name;
  if (!topic) back.delete("topic");

  return <main className={`${styles.page} ${styles.stockPage}`} style={withTokenColors(token)}>
    <Link className={styles.stockBack} to={{ pathname: ROUTER_PATH.hotspots, search: back.toString() }}>
      <ArrowLeftOutlined aria-hidden /> {topic ? `返回${topicName || "题材详情"}` : back.has("board") ? "返回板块详情" : "返回市场热点"}
    </Link>
    {isStockSymbol(symbol)
      ? <StockDetail key={symbol} symbol={symbol} stockName={stockName} topic={topic} topicName={topicName} />
      : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="股票代码无效" />}
  </main>;
}

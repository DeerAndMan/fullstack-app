import { useState } from "react";
import { Alert, Button, Divider, Drawer, Empty, Segmented, Skeleton, Tag, theme } from "antd";
import dayjs from "dayjs";

import { ConceptReason } from "./concept-reason";
import { isStockSymbol, localDate, marketFeeds, marketSession, moneyText, updateTime, useMarketFeed } from "./data";
import { amountColumn, Change, changeColumn, DataTable, nameColumn, withTokenColors } from "./panels";
import { StockLink } from "./stock-link";
import styles from "./hotspots.module.css";

import type { MarketQuery, MarketRow } from "./data";

function failed(query: MarketQuery) {
  return query.isError || query.data?.status === "unavailable";
}

function Retry({ queries, message }: { queries: MarketQuery[]; message: string }) {
  return <Alert
    type="warning"
    showIcon
    message={message}
    action={<Button size="small" loading={queries.some(query => query.isFetching)}
      onClick={() => queries.forEach(query => void query.refetch())}>重试</Button>}
  />;
}

function NewsList({ rows }: { rows: MarketRow[] }) {
  return (
    <ul className={styles.headlines}>
      {rows.map(row => <li key={row.code}>
        <time dateTime={row.date?.replace(" ", "T")}>{row.date?.slice(0, 16)}</time>
        <a href={row.url} target="_blank" rel="noreferrer">{row.name}</a>
      </li>)}
    </ul>
  );
}

function memberStatus(row: MarketRow) {
  if (row.change === undefined) return "暂无报价";
  if (row.change > 0) return "上涨";
  if (row.change < 0) return "下跌";
  return "平盘";
}

const memberReasonMissing = "新浪财经收录，暂未查到具体关联出处。";

function TopicContent({ concept, stale }: { concept: MarketRow; stale: boolean }) {
  const [stockFilter, setStockFilter] = useState<"all" | "streak">("all");
  const [memberFilter, setMemberFilter] = useState("全部");
  const indices = useMarketFeed("indices");
  const date = marketSession(indices.data?.rows);
  const members = useMarketFeed("concept-members", { symbol: concept.code });
  const pool = useMarketFeed("limit-up", { date });
  const news = useMarketFeed("concept-news", { symbol: concept.code });
  const memberRows = members.data?.rows ?? [];
  const codes = new Set(memberRows.map(row => row.code));
  const limitUps = (pool.data?.rows ?? []).filter(row => codes.has(row.code))
    .sort((a, b) => (b.count ?? 0) - (a.count ?? 0) || (b.amount ?? 0) - (a.amount ?? 0));
  const streaks = limitUps.filter(row => (row.count ?? 0) >= 2);
  const rising = memberRows.filter(row => (row.change ?? 0) > 0).length;
  const reported = memberRows.filter(row => row.change !== undefined).length;
  const recentStart = dayjs(localDate()).subtract(6, "day").format("YYYY-MM-DD");
  const recentNews = (news.data?.rows ?? []).filter(row => row.date && row.date.slice(0, 10) >= recentStart);
  const olderNews = (news.data?.rows ?? []).filter(row => row.date && row.date.slice(0, 10) < recentStart);
  const stockPending = members.isPending || pool.isPending;
  const stockErrors = [members, pool].filter(failed);
  const stockReady = !stockPending && !stockErrors.length;
  const partial = Boolean(members.data?.message) || (pool.data?.total ?? 0) > (pool.data?.rows.length ?? 0);
  const displayedStocks = stockFilter === "streak" ? streaks : limitUps;
  const leading = memberRows.find(row => row.name === concept.detail);
  const currentLimits = pool.data?.status === "ready" && members.data?.status === "ready"
    ? new Map(limitUps.map(row => [row.code, row.count ?? 1])) : new Map<string, number>();
  const rankedMembers = memberRows.filter(row => memberFilter === "全部" || memberStatus(row) === memberFilter)
    .map(row => {
      const count = currentLimits.get(row.code);
      return { ...row, detail: count === undefined ? memberStatus(row) : count >= 2 ? `${count} 连板` : "涨停" };
    })
    .sort((a, b) => (b.change ?? -Infinity) - (a.change ?? -Infinity) || (b.amount ?? 0) - (a.amount ?? 0));

  const stockNameColumn = { ...nameColumn, width: 122, render: (row: MarketRow) => <StockLink row={row} /> };

  return <>
    <section className={styles.topicSection} aria-label="为什么热">
      <h3>为什么热</h3>
      {stale && <Alert type="warning" showIcon message="题材排行使用最近一次可用快照。" />}
      <dl className={styles.topicFacts}>
        <div><dt>题材涨幅</dt><dd><Change value={concept.change} /></dd></div>
        <div><dt>领涨股</dt><dd>{concept.detail
          ? <StockLink row={leading ?? { ...concept, name: concept.detail }} /> : "—"}</dd></div>
        <div><dt>板块成交额</dt><dd>{moneyText(concept.amount)}</dd></div>
        {reported > 0 && <div><dt>上涨成分股</dt><dd>{rising} / {reported}</dd></div>}
        {stockReady && <div><dt>{partial ? "已匹配涨停 / 连板" : "涨停 / 连板"}</dt><dd>{limitUps.length} / {streaks.length}</dd></div>}
      </dl>
      {recentNews.length > 0
        ? <p className={styles.topicContext}>近期消息线索 <a href={recentNews[0].url} target="_blank" rel="noreferrer">{recentNews[0].name}</a>。消息与走势的因果关系仍需结合原文判断。</p>
        : !news.isPending && !failed(news) && <p className={styles.note}>近 7 天未检索到匹配消息，本轮异动暂无可核验的消息线索。</p>}
      <p className={styles.muted}>板块行情、成分股来自新浪财经。上涨占比只统计有涨跌幅的成分股。</p>
    </section>

    <Divider />

    <section className={styles.topicSection} aria-label="相关股票">
      <div className={styles.sectionHeader}>
        <h3>相关股票</h3>
        {members.data && members.data.status !== "unavailable" &&
          <span className={styles.muted}>已取得 {memberRows.length} 只 · {updateTime(members.data.fetchedAt)}（北京时间）</span>}
      </div>
      {members.isPending && <Skeleton active title={false} paragraph={{ rows: 4 }} />}
      {failed(members) && <Retry queries={[members]} message="暂时无法取得该题材的成分股。" />}
      {members.data?.status === "stale" && <Alert type="warning" showIcon message="当前显示最近一次可用的成分股快照。" />}
      {members.data?.message && !failed(members) && <Alert type="info" showIcon message={members.data.message} />}
      {memberRows.length > 0 && <>
        <Segmented aria-label="筛选相关股票" className={styles.control} value={memberFilter} onChange={setMemberFilter}
          options={["全部", "上涨", "下跌", "平盘"].map(value => ({
            value, label: `${value} ${value === "全部" ? memberRows.length : memberRows.filter(row => memberStatus(row) === value).length}`,
          }))} />
        <DataTable label={`${concept.name}相关股票`} rows={rankedMembers} search searchPlaceholder="搜索股票名称或代码" pageSize={8}
          columns={[stockNameColumn, {
            title: "关联原因", render: row => (
              <div className={styles.conceptReason}>
                {isStockSymbol(row.symbol)
                  ? <ConceptReason concept={concept.code} symbol={row.symbol} missingMessage={memberReasonMissing} />
                  : <span className={styles.muted}>{memberReasonMissing}</span>}
              </div>
            ),
          }, { title: "代码", field: "code", width: 75 },
            { title: "最新价", field: "price", width: 80 }, changeColumn, amountColumn,
            { title: "状态", field: "detail", width: 78, render: row => (
              <Tag color={currentLimits.has(row.code) || (row.change ?? 0) > 0 ? "red" : (row.change ?? 0) < 0 ? "green" : undefined}>
                {row.detail}
              </Tag>
            ) }]} />
      </>}
      {members.data?.status === "ready" && !memberRows.length &&
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="行情源未列出该题材的成分股。" />}
    </section>

    <Divider />

    <section className={styles.topicSection} aria-label="涨停与连板股">
      <h3>涨停与连板股</h3>
      <p className={styles.note}>数据日 {date} · 新浪财经成分股与东方财富涨停池按股票代码匹配，连续 2 板起计为连板。</p>
      {stockPending && <Skeleton active title={false} paragraph={{ rows: 3 }} />}
      {!stockPending && stockErrors.length > 0 && <Retry queries={stockErrors} message={
        pool.data?.message || members.data?.message || "暂无法核对成分股与涨停池。"
      } />}
      {stockReady && <>
        {(members.data?.status === "stale" || pool.data?.status === "stale") &&
          <Alert type="warning" showIcon message="部分数据来自最近一次可用快照。" />}
        {partial && <Alert type="info" showIcon message={members.data?.message || "涨停池返回不完整，结果仅覆盖已取得的股票。"} />}
        <Segmented className={styles.control} value={stockFilter} onChange={setStockFilter}
          options={[{ label: `全部涨停 ${limitUps.length}`, value: "all" }, { label: `连板 ${streaks.length}`, value: "streak" }]} />
        {displayedStocks.length > 0 ? <DataTable label={`${concept.name}涨停与连板`} rows={displayedStocks}
          columns={[{ title: "代码", field: "code" }, {
            ...stockNameColumn,
            render: row => <StockLink row={{ ...row, symbol: row.symbol || memberRows.find(member => member.code === row.code)?.symbol }} />,
          }, { title: "连板数", field: "count", width: 95 },
            changeColumn, { title: "成交额", field: "amount", money: true }]} />
          : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={
            !memberRows.length ? "数据源未列出该题材成分股。" :
              stockFilter === "streak" ? "未匹配到连板股。" : "未匹配到涨停股。"
          } />}
      </>}
      <a className={styles.source} href={marketFeeds["limit-up"]} target="_blank" rel="noreferrer">查看东方财富涨停池</a>
    </section>

    <Divider />

    <section className={styles.topicSection} aria-label="相关消息原文">
      <h3>相关消息原文</h3>
      {news.isPending && <Skeleton active title={false} paragraph={{ rows: 3 }} />}
      {failed(news) && <Retry queries={[news]} message={news.data?.message || "消息检索暂不可用。"} />}
      {!failed(news) && news.data?.message && <p className={styles.note}>{news.data.message}</p>}
      {news.data?.status === "stale" && <p className={styles.note}>以下为最近一次可用的检索结果。</p>}
      {recentNews.length > 0 && <NewsList rows={recentNews.slice(0, 6)} />}
      {olderNews.length > 0 && <>
        <h4>更早的背景报道</h4>
        <NewsList rows={olderNews.slice(0, 4)} />
      </>}
      {news.data?.status === "ready" && !news.data.rows.length && <p className={styles.note}>没有找到匹配的消息原文。</p>}
      <a className={styles.source} href={news.data?.sourceUrl ?? `${marketFeeds["concept-news"]}?${new URLSearchParams({ q: concept.name })}`}
        target="_blank" rel="noreferrer">在华尔街见闻搜索该题材</a>
    </section>
  </>;
}

export function TopicDetail({ code, onClose }: { code: string; onClose: () => void }) {
  const { token } = theme.useToken();
  const catalog = useMarketFeed("concept");
  const concept = catalog.data?.rows.find(row => row.code === code);

  return (
    <Drawer open title={concept?.name ?? "题材详情"} onClose={onClose} width="min(100vw, 1040px)"
      styles={{ body: withTokenColors(token) }} rootClassName={styles.topicDrawer}>
      {concept ? <TopicContent concept={concept} stale={catalog.data?.status === "stale"} /> :
        catalog.isPending ? <Skeleton active /> :
          failed(catalog) ? <Retry queries={[catalog]} message="题材目录暂不可用。" /> :
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前题材目录中没有此题材。" />}
    </Drawer>
  );
}

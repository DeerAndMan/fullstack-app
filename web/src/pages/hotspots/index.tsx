import { useState } from "react";
import { Link, Navigate, useSearchParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { Button, Progress, Segmented, Tabs, theme } from "antd";
import { ReloadOutlined, RightOutlined } from "@ant-design/icons";

import { BoardDetail, boardFeeds } from "./board-detail";
import { localDate, marketSession, moneyText, useMarketFeed } from "./data";
import {
  amountColumn, CalendarPanel, Change, changeColumn, DataTable, FeedPanel,
  nameColumn, SessionDate, withTokenColors,
} from "./panels";
import styles from "./hotspots.module.css";
import { Headlines } from "./headlines";
import { displayMarketDate } from "./calendar";
import { MarketHistory } from "./market-history";
import { conceptSources, fundSources, industrySources, limitSources, SourcePicker, useMarketSource } from "./source";
import { StockLink } from "./stock-link";
import { TopicDetail } from "./topic-detail";
import ROUTER_PATH from "@/router/section/router-path";

import type { MarketRow } from "./data";
import type { BoardFeed } from "./board-detail";
import type { ReactNode } from "react";

const priceColumn = { title: "最新价", field: "price" } as const;
const codeColumn = { title: "代码", field: "code", render: (row: MarketRow) => <StockLink row={row} label={row.code} /> } as const;
const stockNameColumn = { ...nameColumn, render: (row: MarketRow) => <StockLink row={row} /> };
const poolLabels = { "limit-up": "涨停股", "opened-limit": "炸板股", "limit-down": "跌停股" } as const;
const structuralBoard = /^(?:(?:昨日|百日)(?:涨停|连板|新高).*|含[A-Z]+股?|本月解禁|ST板块|准ST股|融资融券|科创50|央企50|超大盘|三板精选|业绩预[升降]|未股改|.+重仓)$/i;

function groupLimitUps(rows: MarketRow[]) {
  const groups = new Map<string, MarketRow[]>();
  for (const row of rows) {
    if (!row.detail) continue;
    const group = groups.get(row.detail) ?? [];
    group.push(row);
    groups.set(row.detail, group);
  }
  return [...groups.entries()].sort((a, b) => b[1].length - a[1].length || a[0].localeCompare(b[0], "zh-CN"));
}

function FocusList({ items, empty }: {
  items: { title: string; detail?: string; value: ReactNode; to?: string; href?: string }[];
  empty: string;
}) {
  if (!items.length) return <p className={styles.note}>{empty}</p>;
  return (
    <ul className={styles.focusList}>
      {items.map(({ title, detail, value, to, href }) => {
        const content = <>
          <div className={styles.focusName}><strong>{title}</strong>{detail && <span className={styles.muted}>{detail}</span>}</div>
          <span className={styles.focusValue}>{value}{(to || href) && <RightOutlined aria-hidden />}</span>
        </>;
        return <li key={to ?? title}>
          {href ? <a href={href} target="_blank" rel="noreferrer" className={styles.focusRow}>{content}</a>
            : to ? <Link to={to} className={styles.focusRow} aria-label={`查看${title}详情`}>{content}</Link>
            : <div className={styles.focusRow}>{content}</div>}
        </li>;
      })}
    </ul>
  );
}

function FundFlowNote({ source, region }: { source?: string; region?: boolean }) {
  if (!source) return null;
  const notes: Record<string, string> = {
    新浪财经: "新浪财经行业资金净额口径。",
    同花顺: "同花顺行业净流入口径。",
    东方财富: `东方财富${region ? "地域板块" : "行业"}主力资金估算口径。`,
  };
  return notes[source] ? <p className={styles.note}>{notes[source]}</p> : null;
}

function boardHref(feed: BoardFeed, code: string, params?: URLSearchParams) {
  const next = new URLSearchParams(params);
  next.set("view", "during");
  next.set("boardFeed", feed);
  next.set("board", code);
  next.delete("topic");
  return `?${next}`;
}

function topicHref(code: string, params: URLSearchParams) {
  const next = new URLSearchParams(params);
  next.set("topic", code);
  return `?${next}`;
}

function Overview() {
  const [params] = useSearchParams();
  const [conceptFeed, setConceptFeed] = useMarketSource("conceptSource", conceptSources);
  const [fundFeed, setFundFeed] = useMarketSource("fundSource", fundSources);
  const [limitFeed, setLimitFeed] = useMarketSource("upSource", limitSources);
  const concepts = useMarketFeed(conceptFeed);
  const funds = useMarketFeed(fundFeed);
  const indices = useMarketFeed("indices");
  const date = marketSession(indices.data?.rows);
  const limits = useMarketFeed(limitFeed, { date });
  const thsLimit = limitFeed === "limit-up-ths";
  const viewHref = (view: string) => {
    const next = new URLSearchParams(params);
    next.set("view", view);
    if (view === "after") next.delete("pool");
    return `?${next}`;
  };

  return (
    <div className={styles.splitGrid}>
      <FeedPanel title="领涨题材" feed={conceptFeed} query={concepts} compact
        note={concepts.data?.status === "ready" ? concepts.data.message : undefined}
        extra={<div className={styles.panelActions}>
          <SourcePicker label="领涨题材来源" value={conceptFeed} options={conceptSources} onChange={setConceptFeed} />
          <Link to={viewHref("during")}>查看板块</Link>
        </div>}>
        {rows => <FocusList
          empty="当前暂无符合条件的领涨题材。"
          items={rows.filter(row => (row.change ?? 0) > 0 && !structuralBoard.test(row.name)).slice(0, 4).map(row => ({
            title: row.name, detail: row.detail ? `领涨股 ${row.detail}` : undefined,
            value: <Change value={row.change} />,
            to: conceptFeed === "concept" ? topicHref(row.code, params)
              : conceptFeed === "concept-eastmoney" ? boardHref(conceptFeed, row.code, params) : undefined,
            href: conceptFeed === "concept-ths" ? row.url : undefined,
          }))}
        />}
      </FeedPanel>
      <FeedPanel title="涨停聚集" feed={limitFeed} query={limits} compact
        note={`数据日 ${date} · ${thsLimit ? "同花顺涨停池，按返回顺序展示" : "仅列出至少 2 股涨停的行业"}`}
        extra={<div className={styles.panelActions}>
          <SourcePicker label="涨停聚集来源" value={limitFeed} options={limitSources} onChange={setLimitFeed} />
          <Link to={viewHref("after")}>查看梯队</Link>
        </div>}>
        {rows => thsLimit ? <ul className={styles.focusList}>
          {rows.slice(0, 4).map(row => <li key={row.code}>
            <div className={styles.focusRow}><StockLink row={row} /><span>{row.count ? `${row.count} 板` : row.code}</span></div>
          </li>)}
        </ul> : <FocusList
          empty="当前没有至少 2 股涨停的行业。"
          items={groupLimitUps(rows).filter(([, group]) => group.length >= 2).slice(0, 4).map(([title, group]) => {
            const strongest = [...group].sort((a, b) => (b.count ?? 0) - (a.count ?? 0))[0];
            return {
              title, detail: (strongest.count ?? 0) > 1 ? `${strongest.name} · ${strongest.count} 连板` : strongest.name,
              value: `${group.length} 股涨停`,
            };
          })}
        />}
      </FeedPanel>
      <FeedPanel title="资金动向" feed={fundFeed} query={funds} compact
        note={funds.data?.status === "ready" ? funds.data.message : undefined}
        extra={<div className={styles.panelActions}>
          <SourcePicker label="资金动向来源" value={fundFeed} options={fundSources} onChange={setFundFeed} />
          <Link to={viewHref("during")}>查看资金</Link>
        </div>}>
        {rows => {
          const inflows = rows.filter(row => (row.net ?? 0) > 0).slice(0, 2);
          const outflows = rows.filter(row => (row.net ?? 0) < 0).slice(-2).reverse();
          return <>
            <FocusList empty="当前没有可展示的净流入或净流出。"
              items={[...inflows, ...outflows].map(row => ({
                title: row.name, detail: (row.net ?? 0) > 0 ? "净流入" : "净流出",
                value: moneyText(row.net),
                to: fundFeed === "fund-flow-ths" ? undefined : boardHref(fundFeed, row.code, params),
                href: fundFeed === "fund-flow-ths" ? row.url : undefined,
              }))} />
            <FundFlowNote source={funds.data?.source} />
          </>;
        }}
      </FeedPanel>
      <Headlines compact />
    </div>
  );
}

function Boards() {
  const [params, setParams] = useSearchParams();
  const kind = params.get("boards") === "industry" || params.get("boardFeed")?.startsWith("industry") ? "industry" : "concept";
  const [conceptFeed, setConceptFeed] = useMarketSource("conceptSource", conceptSources);
  const [industryFeed, setIndustryFeed] = useMarketSource("industrySource", industrySources);
  const feed = kind === "industry" ? industryFeed : conceptFeed;
  const query = useMarketFeed(feed);
  return (
    <FeedPanel title="概念与行业板块" feed={feed} query={query}
      extra={<div className={styles.panelActions}><Segmented aria-label="板块类型" value={kind} onChange={value => setParams(previous => {
        const next = new URLSearchParams(previous);
        if (value === "industry") next.set("boards", value);
        else next.delete("boards");
        return next;
      }, { replace: true })} options={[{ label: "行业", value: "industry" }, { label: "概念", value: "concept" }]} />
        {kind === "industry"
          ? <SourcePicker label="行业板块来源" value={industryFeed} options={industrySources} onChange={setIndustryFeed} />
          : <SourcePicker label="概念板块来源" value={conceptFeed} options={conceptSources} onChange={setConceptFeed} />}
      </div>}
      note={query.data?.status === "ready" ? query.data.message : undefined}
    >
      {rows => <DataTable key={feed} label={kind === "industry" ? "行业板块" : "概念板块"} search rows={rows}
        columns={[{
          ...nameColumn,
          render: row => {
            if (feed === "industry-ths" || feed === "concept-ths") return <a href={row.url} target="_blank" rel="noreferrer">{row.name}</a>;
            return <Link to={feed === "concept" ? topicHref(row.code, params) : boardHref(feed, row.code, params)}>{row.name}</Link>;
          },
        }, changeColumn, ...(feed === "concept-ths" ? [{ title: "资金净额", field: "net", money: true } as const] : [amountColumn]), {
          title: "领涨股", field: "detail", width: 125,
          render: row => row.detail ? <StockLink row={{ ...row, name: row.detail }} /> : "—",
        }]} />}
    </FeedPanel>
  );
}

function FundFlow() {
  const [params, setParams] = useSearchParams();
  const region = params.get("fund") === "region" || params.get("boardFeed") === "region-fund-flow";
  const [industryFeed, setIndustryFeed] = useMarketSource("fundSource", fundSources);
  const feed = region ? "region-fund-flow" : industryFeed;
  const selectedDirection = params.get("flow");
  const defaultDirection = region ? "all" : "in";
  const direction = selectedDirection === "out" || selectedDirection === "all" || selectedDirection === "in"
    ? selectedDirection : defaultDirection;
  const query = useMarketFeed(feed);
  const netTitle = query.data?.source === "东方财富" ? "主力净额" : "资金净额";
  const title = region ? "区域资金轮动" : "行业资金轮动";
  return (
    <FeedPanel title={title} feed={feed} query={query}
      note={query.data?.status === "ready" ? query.data.message : undefined}
      extra={<div className={styles.panelActions}><Segmented aria-label="资金板块类型" value={region ? "region" : "industry"} onChange={value => setParams(previous => {
        const next = new URLSearchParams(previous);
        if (value === "region") next.set("fund", value);
        else next.delete("fund");
        return next;
      }, { replace: true })} options={[{ label: "行业", value: "industry" }, { label: "区域", value: "region" }]} />
        {!region && <SourcePicker label="行业资金来源" value={industryFeed} options={fundSources} onChange={setIndustryFeed} />}
      </div>}
      controls={<>
        <Segmented aria-label="资金流向" className={styles.topControls} value={direction} onChange={value => setParams(previous => {
          const next = new URLSearchParams(previous);
          if (value === defaultDirection) next.delete("flow");
          else next.set("flow", value);
          return next;
        }, { replace: true })} options={[{ label: "全部", value: "all" }, { label: "净流入", value: "in" }, { label: "净流出", value: "out" }]} />
        <FundFlowNote source={query.data?.source} region={region} />
      </>}
    >
      {rows => {
        const ranked = direction === "all" ? rows : direction === "in"
          ? rows.filter(row => (row.net ?? 0) > 0)
          : rows.filter(row => (row.net ?? 0) < 0).reverse();
        return <DataTable key={`${feed}:${direction}`} label={title} rows={ranked}
          columns={[
            { ...nameColumn, render: row => feed === "fund-flow-ths"
              ? <a href={row.url} target="_blank" rel="noreferrer">{row.name}</a>
              : <Link to={boardHref(feed, row.code, params)}>{row.name}</Link> },
            { title: netTitle, field: "net", money: true }, changeColumn,
          ]} />;
      }}
    </FeedPanel>
  );
}

function LimitBreakdown({ rows, withIndustries }: { rows: MarketRow[]; withIndustries: boolean }) {
  const counts = new Map<number, number>();
  for (const row of rows) {
    if (row.count) counts.set(row.count, (counts.get(row.count) ?? 0) + 1);
  }
  const tiers = [...counts.entries()].sort((a, b) => b[0] - a[0])
    .map(([tier, count]) => [`${tier} 板`, count] as const);
  const groups = groupLimitUps(rows).map(([name, group]) => [name, group.length] as const);
  return (
    <div className={styles.tierGrid}>
      {[{ title: "连板梯队", ranks: tiers }, ...(withIndustries ? [{ title: "涨停行业", ranks: groups }] : [])].map(({ title, ranks }) => (
        <div key={title}>
          <h3>{title}</h3>
          {ranks.slice(0, 6).map(([name, count]) => <div className={styles.rankBar} key={name}>
            <span>{name}</span>
            <Progress percent={100 * count / rows.length} showInfo={false} size="small" />
            <strong>{count}</strong>
          </div>)}
        </div>
      ))}
    </div>
  );
}

const openedSources = [
  { value: "opened-limit", label: "东方财富" },
  { value: "opened-limit-ths", label: "同花顺" },
] as const;

function LimitPools() {
  const [params, setParams] = useSearchParams();
  const pool = params.get("pool");
  const poolKind = pool === "opened-limit" || pool === "limit-down" ? pool : "limit-up";
  const [upFeed, setUpFeed] = useMarketSource("upSource", limitSources);
  const [openedFeed, setOpenedFeed] = useMarketSource("openedSource", openedSources);
  const feed = poolKind === "limit-up" ? upFeed : poolKind === "opened-limit" ? openedFeed : poolKind;
  const ths = feed === "limit-up-ths" || feed === "opened-limit-ths";
  const selectedDate = params.get("poolDate") ?? undefined;
  const indices = useMarketFeed("indices");
  const date = selectedDate ?? marketSession(indices.data?.rows);
  const query = useMarketFeed(feed, { date });
  return (
    <FeedPanel title="涨停与炸板复盘" feed={feed} query={query}
      extra={<div className={styles.panelActions}>
        {poolKind === "limit-up" && <SourcePicker label="涨停股池来源" value={upFeed} options={limitSources} onChange={setUpFeed} />}
        {poolKind === "opened-limit" && <SourcePicker label="炸板股池来源" value={openedFeed} options={openedSources} onChange={setOpenedFeed} />}
        <SessionDate value={date} onChange={value => setParams(previous => {
        const next = new URLSearchParams(previous);
        if (value) next.set("poolDate", value);
        else next.delete("poolDate");
        return next;
      }, { replace: true })} clearable={Boolean(selectedDate)} /></div>}
      note={ths ? query.data?.message || (poolKind === "limit-up"
        ? "同花顺股池展示股票、代码及来源已披露的连板数。"
        : "同花顺炸板池展示股票和代码。") : undefined}
      controls={<Segmented
        aria-label="股池类型"
        className={styles.topControls}
        value={poolKind}
        onChange={value => setParams(previous => {
          const next = new URLSearchParams(previous);
          if (value === "limit-up") next.delete("pool");
          else next.set("pool", value);
          return next;
        }, { replace: true })}
        options={Object.entries(poolLabels).map(([value, label]) => ({ value: value as keyof typeof poolLabels, label }))}
      />}>
      {rows => (
        <>
          <div className={styles.chartToolbar}>
            <span>{poolLabels[poolKind]} {query.data?.total ?? rows.length} 只</span>
            <span className={styles.muted}>数据日 {date}</span>
          </div>
          {poolKind === "limit-up" && <LimitBreakdown rows={rows} withIndustries={!ths} />}
          <DataTable key={feed} label={`${poolLabels[poolKind]}-${date}`} rows={rows} search columns={[
            codeColumn, stockNameColumn,
            ...(ths ? [] : [{ title: "行业", field: "detail", width: 115 } as const, priceColumn, changeColumn]),
            ...(poolKind === "limit-up" ? [{ title: "连板数", field: "count" } as const] : []),
            ...(ths ? [] : [amountColumn]),
          ]} />
        </>
      )}
    </FeedPanel>
  );
}

const leaderboardSources = [
  { value: "leaderboard", label: "东方财富" },
  { value: "leaderboard-kaipanla", label: "开盘啦" },
] as const;

function Leaderboard() {
  const [params, setParams] = useSearchParams();
  const date = params.get("leaderDate") ?? undefined;
  const [feed, setFeed] = useMarketSource("leaderSource", leaderboardSources);
  const query = useMarketFeed(feed, { date });
  const kpl = feed === "leaderboard-kaipanla";
  return (
    <FeedPanel title="个股龙虎榜" feed={feed} query={query}
      note={kpl ? "开盘啦按股票汇总上榜记录，最多显示 100 只。" : "按披露净买额排序，最多展示 100 条上榜记录。同一股票可能因不同原因多次上榜。"}
      extra={<div className={styles.panelActions}>
        <SourcePicker label="龙虎榜来源" value={feed} options={leaderboardSources} onChange={setFeed} />
        <SessionDate value={date} onChange={value => setParams(previous => {
        const next = new URLSearchParams(previous);
        if (value) next.set("leaderDate", value);
        else next.delete("leaderDate");
        return next;
      }, { replace: true })} clearable /></div>}
    >
      {rows => <DataTable label={`个股龙虎榜-${rows[0].date ?? ""}`} rows={rows} search columns={[
        { title: "交易日", field: "date", width: 105 }, stockNameColumn, changeColumn,
        { title: "龙虎榜净买额", field: "net", money: true, width: 130 },
        kpl ? { title: "上榜次数", field: "count" } : { title: "上榜原因", field: "detail", width: 330 },
      ]} />}
    </FeedPanel>
  );
}

const sections = [
  { key: "overview", label: "热点焦点", content: () => <Overview /> },
  { key: "history", label: "历史冷热", content: () => <MarketHistory /> },
  { key: "during", label: "板块与资金", content: () => <div className={styles.splitGrid}><Boards /><FundFlow /></div> },
  { key: "after", label: "涨停与复盘", content: () => <div className={styles.stack}><LimitPools /><Leaderboard /></div> },
  { key: "before", label: "重点事件", content: () => <div className={styles.stack}><Headlines /><CalendarPanel /></div> },
];

export default function Hotspots() {
  const { token } = theme.useToken();
  const [params, setParams] = useSearchParams();
  const client = useQueryClient();
  const selected = sections.find(section => section.key === params.get("view")) ?? sections[0];
  const topic = params.get("topic");
  const board = params.get("board");
  const boardFeed = boardFeeds.find(feed => feed === params.get("boardFeed"));
  const [refreshing, setRefreshing] = useState(false);
  const refresh = async () => {
    setRefreshing(true);
    try {
      await client.refetchQueries({ queryKey: ["market"], type: "active" });
    } finally {
      setRefreshing(false);
    }
  };
  if (params.get("view") === "research") {
    const next = new URLSearchParams(params);
    next.delete("view");
    next.delete("topic");
    return <Navigate replace to={{ pathname: ROUTER_PATH.research, search: next.toString() }} />;
  }
  return (
    <main className={styles.page} style={withTokenColors(token)}>
      <header className={styles.header}>
        <h1>市场热点</h1>
        <div className={styles.headerActions}>
          <Link to={ROUTER_PATH.research}>题材查询</Link>
          <time dateTime={localDate()}>{displayMarketDate(localDate())} · 北京时间</time>
          <Button type="primary" icon={<ReloadOutlined aria-hidden />} loading={refreshing} onClick={() => void refresh()}>刷新数据</Button>
        </div>
      </header>
      <Tabs
        activeKey={selected.key}
        destroyOnHidden
        items={sections.map(({ key, label, content }) => ({ key, label, children: key === selected.key ? content() : null }))}
        onChange={view => setParams(previous => {
          const next = new URLSearchParams(previous);
          next.delete("topic");
          next.delete("board");
          next.delete("boardFeed");
          if (view === "overview") next.delete("view");
          else next.set("view", view);
          return next;
        })}
      />
      {board && boardFeed ? <BoardDetail key={`${boardFeed}:${board}`} code={board} feed={boardFeed} onClose={() => setParams(previous => {
        const next = new URLSearchParams(previous);
        next.delete("board");
        next.delete("boardFeed");
        return next;
      }, { replace: true })} /> : topic && <TopicDetail key={topic} code={topic} onClose={() => setParams(previous => {
        const next = new URLSearchParams(previous);
        next.delete("topic");
        return next;
      }, { replace: true })} />}
    </main>
  );
}

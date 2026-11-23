import { ReloadOutlined } from "@ant-design/icons";
import { Button, Segmented } from "antd";

import { useMarketFeed } from "./data";
import { FeedPanel } from "./panels";
import { useFlashRefresh } from "./flash-refresh";
import { NewsList } from "./news-list";
import { useRememberedMarketSource } from "./source";
import styles from "./hotspots.module.css";

const sources = [
  { value: "flash-sina", label: "新浪财经" },
  { value: "flash-kaipanla", label: "开盘啦" },
  { value: "flash-eastmoney", label: "东方财富" },
  { value: "flash-ths", label: "同花顺" },
] as const;

const routineIndexQuote = /^(?:香港|美国|日本|欧洲)?(?:A股|港股|美股|三大(?:股指|指数)|沪深(?:两市|三大指数)|沪指|上证(?:指数|综指)|深证成指|创业板指|科创50|沪深300|恒生(?:科技)?指数|纳斯达克(?:指数)?|道琼斯(?:指数)?|标普500(?:指数)?|日经225(?:指数)?)[^。；：]{0,16}(?:涨|跌|高开|低开|开盘|收盘|点位)/;
const routineEtfQuote = /^(?:部分)?(?:宽基|沪深300|中证(?:500|1000|2000)|上证50|科创50|创业板)\s*ETF[^。；：]{0,15}(?:放量|缩量|成交|涨|跌|开盘|收盘)/i;

export function Headlines({ compact = false }: { compact?: boolean }) {
  const [feed, setFeed] = useRememberedMarketSource("flashSource", sources);
  const query = useMarketFeed(feed);
  const { active, afterCutoff, countdown, toggle } = useFlashRefresh(feed, query);
  const countingDown = active && !query.isFetching && typeof countdown === "number";
  let refreshStatus = "自动刷新已暂停";
  if (afterCutoff) refreshStatus = "北京时间 16:00 起停止自动刷新";
  else if (query.isFetching) refreshStatus = "正在刷新…";
  else if (countingDown) refreshStatus = `${countdown} 秒后刷新 · 16:00 停止（北京时间）`;
  else if (active) refreshStatus = countdown === null ? "下次刷新将超过 16:00，已跳过" : "正在安排刷新…";

  return (
    <FeedPanel title="重点快讯" feed={feed} query={query} compact={compact} showRefresh={false}
      note={query.data?.message}
      extra={<Segmented aria-label="快讯来源" size={compact ? "small" : "middle"}
        value={feed} onChange={setFeed} options={[...sources]} />}
      controls={<div className={styles.refreshControls}>
        <span className={`${styles.muted} ${styles.refreshStatus}`} role={countingDown ? "timer" : "status"}>
          {refreshStatus}
        </span>
        <div className={styles.refreshActions}>
          <Button size="small" disabled={afterCutoff} onClick={toggle}>{active ? "暂停自动刷新" : "开启自动刷新"}</Button>
          <Button size="small" icon={<ReloadOutlined aria-hidden />} loading={query.isFetching}
            onClick={() => void query.refetch({ cancelRefetch: false })}>刷新一次</Button>
        </div>
      </div>}
    >
      {rows => {
        const selected = rows.filter(row =>
          !routineIndexQuote.test(row.name) && !routineEtfQuote.test(row.name),
        ).sort((a, b) => (b.date ?? "").localeCompare(a.date ?? "")).slice(0, compact ? 4 : 12);
        if (!selected.length) return <p className={styles.note}>近期暂无重点快讯。</p>;
        return <NewsList rows={selected} />;
      }}
    </FeedPanel>
  );
}

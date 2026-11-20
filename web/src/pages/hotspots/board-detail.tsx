import { Drawer, theme } from "antd";

import { moneyText, useMarketFeed } from "./data";
import { amountColumn, Change, changeColumn, DataTable, FeedPanel, nameColumn, withTokenColors } from "./panels";
import { StockLink } from "./stock-link";
import styles from "./hotspots.module.css";

export const boardFeeds = ["industry", "concept-eastmoney", "industry-eastmoney",
  "fund-flow", "fund-flow-eastmoney", "fund-flow-sina", "region-fund-flow", "region-fund-flow-sina"] as const;
export type BoardFeed = typeof boardFeeds[number];

export function BoardDetail({ code, feed, onClose }: { code: string; feed: BoardFeed; onClose: () => void }) {
  const { token } = theme.useToken();
  const catalog = useMarketFeed(feed);
  const members = useMarketFeed("board-members", { symbol: code });
  const board = catalog.data?.rows.find(row => row.code === code);
  const netLabel = members.data?.source === "新浪财经" ? "资金净额" : "主力净额";
  const summary = members.data?.status !== "unavailable" && members.data?.message;

  return (
    <Drawer open title={board?.name ?? "板块详情"} onClose={onClose}
      width="min(100vw, 1040px)" styles={{ body: withTokenColors(token) }} rootClassName={styles.topicDrawer}>
      {board && <dl className={styles.topicFacts}>
        <div><dt>板块涨跌幅</dt><dd><Change value={board.change} /></dd></div>
        {board.net !== undefined && <div>
          <dt>{catalog.data?.source === "新浪财经" ? "板块资金净额" : "板块主力净额"}</dt>
          <dd>{moneyText(board.net)}</dd>
        </div>}
      </dl>}
      <FeedPanel title="成分股" feed="board-members" query={members} note={summary || undefined}>
        {rows => <DataTable
          label={`${board?.name ?? code}成分股`}
          search
          searchPlaceholder="搜索股票名称或代码"
          rows={rows}
          columns={[
            { title: "代码", field: "code", width: 85 },
            { ...nameColumn, render: row => <StockLink row={row} /> },
            { title: "最新价", field: "price" }, changeColumn, amountColumn,
            ...(rows.some(row => row.net !== undefined) ? [{ title: netLabel, field: "net", money: true } as const] : []),
          ]}
        />}
      </FeedPanel>
    </Drawer>
  );
}

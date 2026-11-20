import { hotspotDatePicker } from "./calendar";
import { useState } from "react";
import {
  Alert, Button, Card, DatePicker, Empty, Input, Skeleton, Table, theme,
} from "antd";
import { DownloadOutlined, ReloadOutlined } from "@ant-design/icons";
import dayjs from "dayjs";

import {
  exportRows, localDate, marketFeeds, moneyText, percentText, priceText, updateTime, useMarketFeed,
} from "./data";
import styles from "./hotspots.module.css";
import { SourcePicker, useMarketSource } from "./source";

import type { CSSProperties, ReactNode } from "react";
import type { MarketFeed, MarketQuery, MarketRow } from "./data";

type Field = keyof MarketRow;

type Column = {
  title: string;
  field?: Field;
  money?: boolean;
  percent?: boolean;
  width?: number;
  render?: (row: MarketRow) => ReactNode;
};

export const changeColumn: Column = { title: "涨跌幅", field: "change", percent: true, width: 98 };
export const amountColumn: Column = { title: "成交额", field: "amount", money: true, width: 118 };
export const nameColumn: Column = { title: "名称", field: "name", width: 170 };

export function Change({ value }: { value?: number }) {
  return (
    <span className={value && value > 0 ? styles.up : value && value < 0 ? styles.down : undefined}>
      {percentText(value)}
    </span>
  );
}

export function FeedPanel({
  title, feed, query, extra, note, children, controls, compact = false, showRefresh = true,
}: {
  title: string;
  feed: MarketFeed;
  query: MarketQuery;
  extra?: ReactNode;
  note?: string;
  controls?: ReactNode;
  children: (rows: MarketRow[]) => ReactNode;
  compact?: boolean;
  showRefresh?: boolean;
}) {
  const data = query.data;
  const available = data && data.status !== "unavailable";
  return (
    <Card title={title} className={styles.panel} extra={extra} size={compact ? "small" : "default"}>
      {controls}
      {note && <p className={styles.note}>{note}</p>}
      {query.isPending ? <Skeleton active title={false} paragraph={{ rows: 4 }} /> : null}
      {query.isError || data?.status === "unavailable" ? (
        <Alert
          type="warning"
          showIcon
          message={query.isError ? "市场数据请求失败" : data?.message || "数据源暂不可用"}
          action={<Button size="small" loading={query.isFetching} onClick={() => void query.refetch()}>重试</Button>}
        />
      ) : null}
      {data?.status === "stale" && <Alert type="warning" showIcon message="上游更新失败，显示最近一次可用数据" />}
      {available && (data.rows.length ? children(data.rows) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前范围暂无数据" />)}
      <div className={styles.source}>
        <a href={data?.sourceUrl ?? marketFeeds[feed]} target="_blank" rel="noreferrer">{data?.source || "查看数据源"}</a>
        {available && <span>获取于 {updateTime(data.fetchedAt)}（北京时间）</span>}
        {showRefresh && <Button aria-label={`刷新${title}`} type="text" size="small" icon={<ReloadOutlined />} loading={query.isFetching} onClick={() => void query.refetch()} />}
      </div>
    </Card>
  );
}

export function DataTable({
  label, rows, columns, search = false, pageSize = 10, searchPlaceholder = "搜索名称、代码或行业",
}: {
  label: string;
  rows: MarketRow[];
  columns: Column[];
  search?: boolean;
  pageSize?: number;
  searchPlaceholder?: string;
}) {
  const [keyword, setKeyword] = useState("");
  const filter = keyword.trim().toLowerCase();
  const visible = rows.filter(row => [row.name, row.code, row.detail].some(value => value?.toLowerCase().includes(filter)));
  const exportColumns = columns.flatMap(({ title, field }) => field ? [{ title, field }] : []);
  return (
    <>
      <div className={styles.tableControls}>
        {search && <Input.Search value={keyword} onChange={event => setKeyword(event.target.value)} allowClear placeholder={searchPlaceholder} aria-label={`搜索${label}`} />}
        <Button
          icon={<DownloadOutlined aria-hidden />}
          disabled={!visible.length}
          onClick={() => exportRows(label, exportColumns.map(c => c.title), visible.map(row => exportColumns.map(c => row[c.field])))}
        >
          导出表格
        </Button>
      </div>
      <Table<MarketRow>
        size="small"
        rowKey="code"
        dataSource={visible}
        scroll={{ x: "max-content" }}
        pagination={visible.length > pageSize ? { pageSize, showSizeChanger: false, hideOnSinglePage: true } : false}
        columns={columns.map(({ title, field, money, percent, width, render }) => ({
          title, dataIndex: field, width, key: field ?? title,
          align: money || percent || field === "price" || field === "count" ? "right" : "left",
          ellipsis: field === "detail",
          sorter: field && (money || percent || field === "price" || field === "count")
            ? (a, b) => Number(a[field] ?? 0) - Number(b[field] ?? 0)
            : undefined,
          render: (value: MarketRow[Field], row: MarketRow) => {
            if (render) return render(row);
            if (percent) return <Change value={Number.isFinite(value) ? Number(value) : undefined} />;
            if (money) return moneyText(typeof value === "number" ? value : undefined);
            if (field === "price") return priceText(typeof value === "number" ? value : undefined);
            if (field === "name" && row.url) {
              return <a href={row.url} target="_blank" rel="noreferrer">{value}</a>;
            }
            return value || value === 0 ? value : "—";
          },
        }))}
      />
    </>
  );
}

export function SessionDate({ value, onChange, clearable = false }: {
  value?: string; onChange: (value?: string) => void; clearable?: boolean;
}) {
  const today = dayjs(localDate());
  return (
    <DatePicker
      {...hotspotDatePicker}
      aria-label="选择数据日期"
      placeholder="最近交易日"
      allowClear={clearable && hotspotDatePicker.allowClear}
      value={value ? dayjs(value) : null}
      minDate={today.subtract(119, "day")}
      maxDate={today}
      onChange={date => onChange(date?.format("YYYY-MM-DD"))}
    />
  );
}

const calendarSources = [
  { value: "calendar", label: "华尔街见闻" },
  { value: "calendar-sina", label: "新浪财经" },
  { value: "calendar-eastmoney", label: "东方财富" },
  { value: "calendar-ths", label: "同花顺" },
] as const;

export function CalendarPanel() {
  const [date, setDate] = useState(localDate);
  const [feed, setFeed] = useMarketSource("calendarSource", calendarSources);
  const query = useMarketFeed(feed, { date });
  const eastmoney = feed === "calendar-eastmoney";
  const ths = feed === "calendar-ths";
  const note = ths ? "同花顺投资日历按日期展示行业活动、指数调整与其他事件。"
    : eastmoney ? "展示与所选 7 天窗口重叠的东方财富财经事件。"
    : "从所选日期起，展示来源标记为 3 级及以上的事件。时间为北京时间。";
  return (
    <FeedPanel title="未来 7 天重点事件" feed={feed} query={query}
      note={query.data?.status === "ready" && query.data.message ? note + " " + query.data.message : note}
      extra={<div className={styles.panelActions}>
        <SourcePicker label="重点事件来源" value={feed} options={calendarSources} onChange={setFeed} />
        <SessionDate value={date} onChange={value => value && setDate(value)} />
      </div>}>
      {rows => ths ? <DataTable key={feed} label="同花顺投资日历" rows={rows} columns={[
        { title: "日期", field: "date", width: 130 }, { title: "事件", field: "name", width: 350 },
      ]} /> : eastmoney ? <DataTable key={feed} label="东方财富财经事件" rows={rows} columns={[
        { title: "日期", field: "date", width: 215 }, { title: "类型", field: "detail", width: 100 },
        { title: "事件", field: "name", width: 300 }, { title: "地点", field: "actual", width: 100 },
      ]} /> : <DataTable key={feed} label="重点财经事件" rows={rows.filter(row => (row.count ?? 0) >= 3)} columns={[
        { title: "北京时间", field: "date", width: 155 }, { title: "地区", field: "detail", width: 84 },
        { title: "事件", field: "name", width: 300 },
        { title: "前值", field: "previous" }, { title: "预期", field: "forecast" }, { title: "公布", field: "actual" },
      ]} />}
    </FeedPanel>
  );
}

export function withTokenColors(token: ReturnType<typeof theme.useToken>["token"]): CSSProperties {
  return {
    "--market-border": token.colorBorderSecondary,
    "--market-secondary": token.colorTextSecondary,
    "--market-accent": token.colorPrimaryText,
    "--market-up": token.colorErrorText,
    "--market-down": token.colorSuccessText,
  } as CSSProperties;
}

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Card, Popover, Table } from "antd";
import type { TableColumnsType } from "antd";
import { z } from "zod";
import { RequestGet, RequestPost } from "@/api/request-schema";
import { ReviewRefreshNotice } from "./review-update";
import styles from "./review.module.css";

const adviceSchema = z.object({
  market: z.string(), buy: z.string(), averageState: z.string(), styleBias: z.string(),
  sectorState: z.string(), model: z.string(), generatedAt: z.string(), rule: z.string(),
});
const timeline = z.array(z.object({
  date: z.string(), batch: z.string(), advice: adviceSchema.optional(), up: z.number(), highs: z.number(), above: z.number(), eligible: z.number(),
  highDelta: z.number().nullable(), averageChange: z.number().nullable(), indexDivergence: z.boolean().nullable(),
  marketAmount: z.number().nullable(), capacityLeaders: z.array(z.string()).nullable(),
}));
type Session = z.infer<typeof timeline>[number];
const missing = <span className={styles.missing} title="该口径尚未归档">—</span>;
const tips = [
  ["先看是否适合开仓", "一年新高不足 20 家，作者通常暂停常规新开仓；从 20 家及以下重返 20 家以上，再找机会。连续普跌中上涨不足 1,000 家且新高增加，先观察逆势品种。"],
  ["再找行业与风格", "容量看百日新高家数，强度看行业二十日涨幅。连续几天比较前排方向，再比较近期新高指数与平均股价，以及指数相对 EMA5、EMA13 的位置。"],
  ["个股等待确认", "上涨中继、EMA5 大小平台、首次 EMA13 回踩或旧支撑企稳，要结合强股辨识度、十字星和板块呼应。预先确定失效条件，持仓时按计划止损或在一致走强时考虑兑现。"],
  ["ETF 单独判断", "沪深 300 的 50 日占比低于 20% 或高于 70% 表示偏冷或偏热。观察指数创新高时占比是否回落。"],
];

const coreTips = [
  ["休息", <>一年新高不足 <b>20 家</b>，说明可选强股太少，作者通常暂停常规新开仓。重新超过 20 家，才恢复筛选。</>],
  ["加入观察", <>连续普跌时，上涨不足 <b>1,000 家</b>，一年新高家数却增加，说明还有股票逆势走强。此时最值得问的是“哪些股票还在创新高，它们是否属于同一方向”。</>],
  ["列入候选", <>连看几天，找反复排在前面的板块。<b>百日新高家数</b>看这个方向有多少强股，<b>二十日涨幅</b>看板块本身涨得强不强。再从中挑成交活跃、前期涨幅突出、能代表该题材的股票。</>],
  ["考虑买入", <>候选股前面已经走强，现在出现分歧。作者关注十字星及其<b>出现的位置</b>。例如前期大阳后横盘 2–4 天，回踩 EMA5 附近仍守住强势位置，同板块也有股票走强。到这里，再明确“跌到哪里说明判断错了、这笔最多承受多少损失”，据此决定是否参与。</>],
  ["持有、减仓或退出", <>已有仓位时，回看当初买入的依据。支撑和板块呼应仍在，可以按计划跟踪；分歧后重新一致走强，作者倾向波段兑现；触及预设止损则退出。文章没有给所有个股规定统一卖价。</>],
] as const;

function AdvicePanel({row}: {row: Session}) {
  const client = useQueryClient();
  const generate = useMutation({
    mutationFn: async () => (await RequestPost("/api/v1/review/advice", {date: row.date, batch: row.batch}, {schema: adviceSchema})).data,
    onSuccess: () => client.invalidateQueries({queryKey: ["review-timeline"]}),
  });
  const advice = row.advice ?? generate.data;
  return <section className={styles.aiPanel} aria-label="归档 AI 建议">
    <strong>{row.date} · AI 复盘建议</strong>
    {advice ? <><p>{advice.market}</p><p><b>买入评估　</b>{advice.buy}</p>
      <small>{advice.model} · {new Date(advice.generatedAt).toLocaleString("zh-CN")} 生成 · 依据当日收盘归档</small></>
      : <><Button loading={generate.isPending} onClick={() => generate.mutate()}>生成并保存当日建议</Button>
        {generate.isError && <Alert type="error" message="AI 建议生成失败，可重试。"/>}</>}
  </section>;
}


function divergence(row: Session) {
  const flags = [];
  if (row.indexDivergence) flags.push("指数／占比背离");
  if (row.up < 1000 && row.highDelta != null && row.highDelta > 0) flags.push("普跌新高增加");
  if (flags.length) return <strong className={styles.signal}>{flags.join(" · ")}</strong>;
  if (row.indexDivergence == null || (row.up < 1000 && row.highDelta == null)) return missing;
  return "无";
}

export default function ReviewTimeline({ date, select }: { date?: string; select: (date: string) => void }) {
  const history = useQuery({
    queryKey: ["review-timeline", date], enabled: !!date,
    select: (rows: Session[]) => [...rows].sort((a, b) => b.date.localeCompare(a.date)),
    queryFn: async ({signal}) => (await RequestGet("/api/v1/review/timeline", {
      params: {through: date}, schema: timeline, signal,
    })).data,
  });
  const current = history.data?.find(row => row.date === date);
  const columns: TableColumnsType<Session> = [
    {title: "日期", dataIndex: "date", fixed: "left", width: 105,
      render: (value: string) => <button className={styles.dateLink} aria-current={value === date ? "date" : undefined} onClick={() => select(value)}>{value}</button>},
    {title: "平均股价状态", render: (_, r) => r.advice?.averageState || missing},
    {title: "风格偏向", render: (_, r) => r.advice?.styleBias || missing},
    {title: "上涨家数", dataIndex: "up", render: (n: number) => <span className={n < 1000 ? styles.cold : undefined}>{n}</span>},
    {title: "平均股价涨跌", render: (_, r) => r.averageChange == null ? missing :
      <span className={r.averageChange > 0 ? styles.gain : r.averageChange < 0 ? styles.loss : undefined}>{r.averageChange.toFixed(2)}%</span>},
    {title: "背离观察", render: (_, r) => divergence(r)},
    {title: "一年新高", render: (_, r) => <span className={r.highs < 20 ? styles.cold : styles.highs}>{r.highs}</span>},
    {title: "含北交", render: () => <span className={styles.missing} title="缺少北交所同口径的一年新高行情">—</span>},
    {title: "成交额（亿）", render: (_, r) => r.marketAmount == null ? missing : (r.marketAmount / 1e8).toFixed(0)},
    {title: <>核心强势板块<small className={styles.leaders}>容量前三参考</small></>, width: 205,
      render: (_, r) => r.capacityLeaders == null ? missing : r.capacityLeaders.join("、") || "无"},
    {title: "板块状态 · AI", render: (_, r) => r.advice?.sectorState || missing},
    {title: "沪深300 50日占比", width: 110, render: (_, r) => {
      if (!r.eligible) return missing;
      const ratio = 100 * r.above / r.eligible;
      return <span className={ratio < 20 ? styles.cold : ratio > 70 ? styles.hot : undefined} title={r.above + " / " + r.eligible}>{ratio.toFixed(2)}%</span>;
    }},
  ];
  return <Card title="逐日复盘" extra={<div className={styles.actions}><Popover trigger="click" placement="bottomRight" title="核心提示"
    content={<div className={styles.coreTips}><table><thead><tr><th>你要判断的操作</th><th>按作者的思路，什么情况下走到这一步</th></tr></thead>
      <tbody>{coreTips.map(([label, detail]) => <tr key={label}><th scope="row">{label}</th><td>{detail}</td></tr>)}</tbody></table></div>}>
    <Button size="small">TIP · 核心提示</Button></Popover><Popover trigger="click" placement="bottomRight" title="原文操作要点"
    content={<div className={styles.tipBody}>{tips.map(([title,body]) => <p key={title}><b>{title}</b><br/>{body}</p>)}</div>}>
    <Button size="small" aria-label="TIP · 查看原文操作要点">TIP · 操作要点</Button></Popover></div>} className={styles.matrix}>
    <ReviewRefreshNotice displayDate={current?.date}/>
    {current && <AdvicePanel key={current.batch} row={current}/>}
    {history.isError ? <Alert type="error" message="逐日归档读取失败" action={<Button onClick={() => void history.refetch()}>重试</Button>}/> :
      <Table<Session> size="small" bordered rowKey="date" columns={columns} dataSource={history.data ?? []}
        pagination={false} loading={history.isFetching} scroll={{x: 1240}}
        locale={{emptyText: "暂无已发布复盘"}} />}
    <p className={styles.caption}>截至所选日期的最近 15 个归档日。点击日期查看当天名单。核心板块栏展示容量前三作为参考，按百日新高家数排列，同数按代码排序。平均股价状态和风格偏向由归档均线与指数涨跌幅核对，板块状态为 AI 解读。横线表示缺少依据或尚未生成。</p>
  </Card>;
}

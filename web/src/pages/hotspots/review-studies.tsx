import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Card, Empty, Input, InputNumber, Select, Space, Spin, Table } from "antd";
import { z } from "zod";
import styles from "./review.module.css";
import { RequestGet, RequestPut } from "@/api/request-schema";

const styleData = z.object({
  spread: z.number(),
  index: z.object({close:z.number(),previous:z.number(),previous20High:z.number(),divergence:z.boolean().nullable()}).optional(),
  rows: z.array(z.object({ code: z.string(), close: z.number(), previous: z.number(), ema5: z.number(), ema13: z.number(), change: z.number() })),
});
const boards = z.object({
  base: z.string(), unclassified: z.array(z.string()),
  rows: z.array(z.object({ code: z.string(), name: z.string(), eligible: z.number(), members: z.number(), hits: z.array(z.string()), strength: z.number().nullable(), strengthMethod: z.string().optional(), change:z.number().nullish(), amount:z.number().nullish(), start: z.number().nullable(), close: z.number().nullable() })),
});
const patterns = z.object({
  total: z.number(), rows: z.array(z.object({ symbol: z.string(), name: z.string(), kind: z.string(), anchor: z.string(), days: z.number(), big: z.number(), near: z.number(), width: z.number(), doji: z.number(), shrink: z.number(), close: z.number(), high: z.number(), open: z.number(), low: z.number(), ema5: z.number(), ema13: z.number(), amount: z.number(), signalDate: z.string(), triggerDate: z.string().optional() })),
});
type Day = { date: string; batch: string };
type Params = Record<string, string | number>;
function useStudy<T extends z.ZodTypeAny>(day: Day, kind: string, schema: T, params: Params = {}) {
  return useQuery({ queryKey: ["review-study", day.batch, kind, params],
    queryFn: async ({signal}) => (await RequestGet("/api/v1/review/study", {
      params: {date: day.date, batch: day.batch, kind, ...params}, schema: schema.nullable(), signal,
    })).data as z.infer<T> | null });
}
const decimal = (v: number | null, digits = 2) => v == null ? "—" : v.toFixed(digits);
function Status({ loading, failed }: {loading: boolean; failed: boolean}) {
  return failed ? <Alert type="error" message="读取归档失败" /> : loading ? <Spin /> : <Empty description="本日此项尚未归档" />;
}
export function StyleCard({day}: {day: Day}) {
  const q = useStudy(day, "style", styleData);
  return <Card title="风格指数">{q.data ? <>
    <p>近期新高相对平均股价的当日涨跌幅差为 <strong>{decimal(q.data.spread)} 个百分点</strong>。</p>
    {q.data.index && <p>沪深 300 收盘 {decimal(q.data.index.close)}，此前 20 日最高收盘 {decimal(q.data.index.previous20High)}。{q.data.index.divergence===null?"前一交易日占比缺失。":q.data.index.divergence?"指数创新高且站上 MA50 的成份占比回落。建议对照逐日表，跟踪占比回落是否持续。":"当日未出现指数创新高且均线上方占比回落的背离。"}</p>}
    <Table size="small" rowKey="code" pagination={false} scroll={{x: 580}} dataSource={q.data.rows} columns={[
      {title:"指数",dataIndex:"code",render:(v: string)=>v==="880865"?"近期新高 · 880865":"平均股价 · 880003"},
      ...[["收盘","close"],["前收","previous"],["EMA5","ema5"],["EMA13","ema13"],["涨跌幅（%）","change"]].map(([title,dataIndex])=>({title,dataIndex,render:(n:number)=>decimal(n)})),
    ]}/></> : <Status loading={q.isPending} failed={q.isError}/>}</Card>;
}
export function SectorCard({day}: {day: Day}) {
  const q = useStudy(day, "sectors", boards);
  const rows = [...(q.data?.rows ?? [])].sort((a,b) => b.hits.length - a.hits.length || a.code.localeCompare(b.code));
  const strengths = rows.flatMap(row => row.strength == null ? [] : [row.strength]).sort((a,b) => b-a);
  const cutoff = strengths[Math.min(2, strengths.length-1)];
  const overlap = rows.slice(0, 3).filter(row => row.hits.length > 0 && row.strength != null && row.strength >= cutoff);
  return <Card title={day.date + " · 行业容量与强度"} className={styles.sector}>
    {q.data && <p><strong>行业观察　</strong>{!strengths.length ? "当日缺少二十日涨幅，先按容量核对百日新高名单。"
      : overlap.length ? overlap.map(row => row.name).join("、") + " 同时进入容量与已归档强度前三，优先展开名单，再对照前几个交易日的排名。"
      : "容量前三与已归档的强度前三没有重合，分别查看新高股集中在哪些行业、哪些行业指数涨幅靠前。"}</p>}
    {q.data && <p className={styles.caption}>同花顺行业 · 按百日新高家数排列。蓝框标出有数据行业中二十日涨幅前三（含并列）。展开可查新高名单。</p>}
    {q.data ? <Table size="small" bordered rowKey="code" scroll={{x: 800}}
      pagination={{pageSize:15,showSizeChanger:false,hideOnSinglePage:true}} dataSource={rows.map((row,i) => ({...row,rank:i+1}))}
      expandable={{expandedRowRender:r=><p>新高名单　{r.hits.join("、")||"无"}<br/>有效样本 {r.eligible} / 成份 {r.members}。
        强度基准日 {q.data?.base}。{r.strengthMethod === "dailyReturns" ? "按 20 个交易日公布涨跌幅连乘。" : <>起点收盘 {decimal(r.start,3)}，当日收盘 {decimal(r.close,3)}。</>}</p>}}
      columns={[
        {title:"容量序",dataIndex:"rank",sorter:(a,b)=>a.rank-b.rank},
        {title:"代码",dataIndex:"code"}, {title:"名称",dataIndex:"name"},
        {title:"百日新高（家）",render:(_,r)=>r.hits.length},
        {title:"涨跌幅",render:(_,r)=>r.change == null ? "—" : <span className={r.change>0?styles.gain:r.change<0?styles.loss:undefined}>{decimal(r.change)}%</span>},
        {title:"收盘",render:(_,r)=>decimal(r.close)},
        {title:"总金额（亿）",render:(_,r)=>r.amount == null ? "—" : decimal(r.amount/1e8,1)},
        {title:"二十日涨幅",render:(_,r)=> r.strength == null ? "—" :
          <span className={r.strength>=cutoff?styles.topStrength:undefined}>{decimal(r.strength)}%</span>,
          sorter:(a,b,order)=>a.strength==null || b.strength==null
            ? (Number(a.strength==null)-Number(b.strength==null))*(order==="descend"?-1:1)
            : a.strength-b.strength},
      ]}/> : <Status loading={q.isPending} failed={q.isError}/>}
    {q.data && <p className={styles.caption}>行业归属未覆盖的新高股票 {q.data.unclassified.length} 家。二十日涨幅已换算为百分数，原图 0.12 对应 12%。</p>}
  </Card>;
}
const controls = [
  ["big","大阳实体涨幅下限（%）",1,30,5],["near","贴线或支撑距离上限（%）",0,25,2],
  ["width","平台宽度上限（%）",0,50,8],["doji","十字实体占振幅上限（%）",0,100,35],["shrink","振幅比上限",0,3,.75],
] as const;
const defaults: Record<string, number> = Object.fromEntries(controls.map(([key,,,,v])=>[key,v]));
const names = ["上涨中继","EMA5 小平台","EMA5 大平台","EMA13 回踩","超跌回踩"];
const noteSchema = z.object({text:z.string(),updatedAt:z.string()});
function PatternNote({date,symbol}:{date:string;symbol:string}) {
  const key=["review-note",date,symbol];
  const client=useQueryClient();
  const [draft,setDraft]=useState<string>();
  const note=useQuery({queryKey:key,queryFn:async({signal})=>(await RequestGet("/api/v1/review/note",
    {params:{date,symbol},schema:noteSchema.nullable(),signal})).data});
  const save=useMutation({mutationFn:async(text:string)=>(await RequestPut("/api/v1/review/note",{date,symbol,text},{schema:noteSchema})).data,
    onSuccess:data=>{client.setQueryData(key,data);setDraft(undefined);}});
  return <div><label>个人题材备注<Input.TextArea maxLength={1000} value={draft??note.data?.text??""}
    disabled={note.isPending||note.isError} onChange={e=>setDraft(e.target.value)}/></label>
    <Space><Button loading={save.isPending} disabled={draft===undefined} onClick={()=>save.mutate(draft??"")}>保存备注</Button>
      {note.data && <small>填写于 {new Date(note.data.updatedAt).toLocaleString("zh-CN")}</small>}
      {(save.isError||note.isError)&&<small>备注读写失败，请重试。</small>}</Space></div>;
}
export function PatternCard({day}: {day: Day}) {
  const [params,setParams]=useSearchParams();
  const thresholds = {...defaults};
  for (const [key,,min,max] of controls) {
    const value = params.get(key);
    if (value && Number.isFinite(Number(value)) && Number(value)>=min && Number(value)<=max) thresholds[key]=Number(value);
  }
  const setThresholds = (next:Record<string,number>) => {
    const updated = new URLSearchParams(params);
    for (const [key] of controls) {
      if (next[key]===defaults[key]) updated.delete(key);
      else updated.set(key,String(next[key]));
    }
    setParams(updated,{replace:true});
  };
  const pattern=params.get("pattern")??""; const search=params.get("search")??"";
  const stage=params.get("stage")==="trigger"?"trigger":"signal"; const [page,setPage]=useState(1);
  function filter(key:string, value:string) { const next=new URLSearchParams(params); next.set(key,value); setPage(1); setParams(next,{replace:true}); }
  const q=useStudy(day,"patterns",patterns,{...thresholds,pattern,search,stage,page});
  return <Card title="回踩形态候选">
    <p>下方为首版筛选阈值。调整后按当天归档证据重算。触发采用下一交易日最高价严格突破候选日最高价，两日价格统一复权。</p>
    <Space wrap align="end">{controls.map(([key,label,min,max])=><label key={key}>{label}<br/>
      <InputNumber aria-label={label} value={thresholds[key]} min={min} max={max} step={key==="shrink"?.05:.5} onChange={v=>{if(v!=null && v>=min && v<=max){setPage(1);setThresholds({...thresholds,[key]:v});}}}/></label>)}
      <Button onClick={()=>{setThresholds(defaults);setPage(1);}}>恢复初值</Button></Space>
    <p><Space wrap>
      <Select aria-label="阶段" value={stage} options={[{label:"当日候选",value:"signal"},{label:"当日触发",value:"trigger"}]} onChange={v=>{filter("stage",v);}}/>
      <Select aria-label="形态" value={pattern} options={["",...names].map(value=>({value,label:value||"全部形态"}))} onChange={v=>{filter("pattern",v);}}/>
      <Input.Search key={search} defaultValue={search} placeholder="代码或名称" allowClear onSearch={v=>filter("search",v)}/>
    </Space></p>
    {q.data ? <Table size="small" scroll={{x:820}} rowKey={r=>r.symbol+r.kind+r.anchor} dataSource={q.data.rows}
      pagination={{current:page,pageSize:50,total:q.data.total,onChange:setPage,showSizeChanger:false,showTotal:n=>n+" 个形态信号"}}
      expandable={{expandedRowRender:r=><><p>候选日开 {decimal(r.open)}、高 {decimal(r.high)}、低 {decimal(r.low)}、收 {decimal(r.close)}，成交额 {decimal(r.amount/1e8)} 亿元。EMA5 为 {decimal(r.ema5)}，EMA13 为 {decimal(r.ema13)}。间隔 {r.days} 个交易日，平台宽度 {decimal(r.width)}%，距参考线或支撑 {decimal(r.near)}%，实体占振幅 {decimal(r.doji)}%，振幅比 {decimal(r.shrink)}。</p><PatternNote date={r.signalDate||day.date} symbol={r.symbol}/></>}}
      columns={[{title:"股票",render:(_,r)=>r.name+" "+r.symbol},{title:"形态",dataIndex:"kind"},{title:"大阳日",dataIndex:"anchor"},
        {title:"大阳实体（%）",dataIndex:"big",render:(v:number)=>decimal(v)},{title:"候选日",dataIndex:"signalDate"},{title:"触发日",dataIndex:"triggerDate",render:(v?:string)=>v||"待观察"}]}/>
    : <Status loading={q.isPending} failed={q.isError}/>}</Card>;
}

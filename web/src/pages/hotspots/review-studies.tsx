import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Card, Empty, Input, InputNumber, Select, Space, Spin, Table } from "antd";
import { z } from "zod";
import { RequestGet, RequestPut } from "@/api/request-schema";

const styleData = z.object({
  spread: z.number(),
  index: z.object({close:z.number(),previous:z.number(),previous20High:z.number(),divergence:z.boolean().nullable()}).optional(),
  rows: z.array(z.object({ code: z.string(), close: z.number(), previous: z.number(), ema5: z.number(), ema13: z.number(), change: z.number() })),
});
const boards = z.object({
  base: z.string(), unclassified: z.array(z.string()),
  rows: z.array(z.object({ code: z.string(), name: z.string(), eligible: z.number(), members: z.number(), hits: z.array(z.string()), strength: z.number().nullable(), strengthMethod: z.string().optional(), start: z.number().nullable(), close: z.number().nullable() })),
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
function StyleCard({day}: {day: Day}) {
  const q = useStudy(day, "style", styleData);
  return <Card title="风格指数">{q.data ? <>
    <p>近期新高相对平均股价的当日涨跌幅差为 <strong>{decimal(q.data.spread)} 个百分点</strong>。</p>
    {q.data.index && <p>沪深 300 收盘 {decimal(q.data.index.close)}，此前 20 日最高收盘 {decimal(q.data.index.previous20High)}。{q.data.index.divergence===null?"前一交易日占比缺失。":q.data.index.divergence?"指数创新高且站上 MA50 的成份占比回落。":"当日未出现指数创新高且均线上方占比回落的背离。"}</p>}
    <Table size="small" rowKey="code" pagination={false} scroll={{x: 580}} dataSource={q.data.rows} columns={[
      {title:"指数",dataIndex:"code",render:(v: string)=>v==="880865"?"近期新高 · 880865":"平均股价 · 880003"},
      ...[["收盘","close"],["前收","previous"],["EMA5","ema5"],["EMA13","ema13"],["涨跌幅（%）","change"]].map(([title,dataIndex])=>({title,dataIndex,render:(n:number)=>decimal(n)})),
    ]}/></> : <Status loading={q.isPending} failed={q.isError}/>}</Card>;
}
function SectorCard({day}: {day: Day}) {
  const q=useStudy(day,"sectors",boards);
  return <Card title="行业容量与强度">{q.data ? <>
    <p>容量取百日新高家数，强度取相对 {q.data.base} 收盘的涨跌幅。展开行业可查命中名单。归属未覆盖的新高股票 {q.data.unclassified.length} 家。</p>
    <Table size="small" rowKey="code" scroll={{x:620}} pagination={{pageSize:8,showSizeChanger:false}} dataSource={q.data.rows}
      expandable={{expandedRowRender:r=><p>新高名单　{r.hits.join("、")||"无"}<br/>{r.strengthMethod === "dailyReturns" ? "强度按该行业 20 个交易日的公布涨跌幅连乘，精度受日涨跌幅舍入影响。" : <>起点收盘 {decimal(r.start,3)}　当日收盘 {decimal(r.close,3)}</>}</p>}}
      columns={[
        {title:"行业",dataIndex:"name"}, {title:"百日新高",render:(_,r)=>r.hits.length,sorter:(a,b)=>a.hits.length-b.hits.length},
        {title:"有效样本 / 成份",render:(_,r)=>r.eligible+" / "+r.members},
        {title:"20 日强度（%）",dataIndex:"strength",render:(v:number|null)=>v===null?"缺少指数收盘":decimal(v),sorter:(a,b)=>(a.strength??-Infinity)-(b.strength??-Infinity)},
      ]}/></> : <Status loading={q.isPending} failed={q.isError}/>}</Card>;
}
const controls = [
  ["big","大阳实体涨幅下限（%）",1,30,5],["near","贴线或支撑距离上限（%）",0,25,2],
  ["width","平台宽度上限（%）",0,50,8],["doji","十字实体占振幅上限（%）",0,100,35],["shrink","振幅比上限",0,3,.75],
] as const;
const defaults = Object.fromEntries(controls.map(([key,,,,v])=>[key,v]));
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
function PatternCard({day}: {day: Day}) {
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
export default function ReviewStudies({day}: {day: Day}) {
  return <><StyleCard day={day}/><SectorCard day={day}/><PatternCard day={day}/></>;
}

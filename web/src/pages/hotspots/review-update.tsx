import { useEffect, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button } from "antd";
import { z } from "zod";

import { RequestGet, RequestPost } from "@/api/request-schema";
import styles from "./review.module.css";

const runSchema = z.object({
  provider: z.string().optional(), corePublished: z.boolean().optional(), date: z.string(), id: z.number(), state: z.enum(["", "running", "failed", "done", "partial", "unpublished"]), message: z.string(),
});
const providerSchema = z.object({id: z.string(), label: z.string(), reason: z.string()});
const statusSchema = z.object({ today: z.string(), providers: z.array(providerSchema).optional(), auto: z.boolean().optional(), run: runSchema });

export function useRefreshStatus() {
  return useQuery({
    queryKey: ["review-refresh-status"],
    queryFn: async ({signal}) => (await RequestGet("/api/v1/review/refresh", {schema: statusSchema, signal})).data,
    refetchInterval: query => query.state.data?.run.state === "running" ? 3_000 : query.state.data?.auto ? 20_000 : false,
    refetchOnMount: false,
    retry: false,
  });
}

export function ReviewRefreshNotice({displayDate}: {displayDate?: string}) {
  const {data} = useRefreshStatus();
  const run = data?.run;
  if (!run?.state) return null;
  if (run.state === "unpublished") {
    return <Alert showIcon type="info" className={styles.refreshNotice}
      message={run.date + " · 收盘行情待发布"}
      description={<>{run.message}。{displayDate && <>当前显示 {displayDate} 的归档。</>}数据发布后可再次更新。</>}/>;
  }
  const viewing = displayDate ? "当前显示 " + displayDate + " 的归档。" : "";
  return <Alert showIcon className={styles.refreshNotice}
    type={run.state === "failed" ? "warning" : run.state === "running" || run.state === "partial" ? "info" : "success"}
    message={run.date + (run.corePublished && run.state !== "done" ? " 核心复盘已发布" :
      run.state === "failed" ? " 更新未完成" : run.state === "running" ? " 正在更新" : " 更新完成")}
    description={run.message + "　" + viewing}/>;
}

export default function ReviewUpdate({selectDate, requestedDate}: {selectDate: (date: string) => void; requestedDate?: string}) {
  const client = useQueryClient();
  const [chosen, setChosen] = useState<{page?: string; date: string}>();
  const [source, setSource] = useState("");
  const watching = useRef<number | undefined>(undefined);
  const observed = useRef("");
  const status = useRefreshStatus();
  const selected = (chosen?.page === requestedDate && chosen?.date) || requestedDate || status.data?.today || "";
  const sources = (status.data?.providers ?? [])
    .filter(p => !p.reason && (p.id !== "public" || selected === status.data?.today));
  const provider = sources.find(p => p.id === source)?.id ?? sources[0]?.id;
  const start = useMutation({
    mutationFn: async () => (await RequestPost("/api/v1/review/refresh", {date: selected, provider}, {schema: runSchema})).data,
    onSuccess: run => {
      watching.current = run.id;
      client.setQueryData<z.infer<typeof statusSchema>>(["review-refresh-status"], old => old && {...old, run});
    },
  });
  const run = status.data?.run;
  useEffect(() => {
    if (!run?.state || (run.state === "running" && !run.corePublished)) return;
    const event = `${run.id}:${run.state}:${run.message}`;
    if (observed.current === event || (!run.corePublished && watching.current !== run.id)) return;
    observed.current = event;
    void client.invalidateQueries({predicate: q => typeof q.queryKey[0] === "string" &&
      ["review-dates","review-day","review-timeline","review-study","review-members"].includes(q.queryKey[0])});
    if (watching.current === run.id) {
      watching.current = undefined;
      if (run.corePublished || run.state === "done") selectDate(run.date);
    }
  }, [run, client, selectDate]);

  const busy = start.isPending || run?.state === "running";
  const supplementing = run?.state === "running" && run.corePublished;
  const progress = run?.message.startsWith(run.date) ? run.message : run?.date + " · " + run?.message;
  return <div className={styles.updateBox}>
    <div className={styles.actions}>
      <label>日线来源　<select aria-label="日线来源" value={provider ?? ""} disabled={busy}
        onChange={e => setSource(e.target.value)}>{sources.map(p =>
        <option key={p.id} value={p.id}>{p.label}</option>
      )}</select></label>
      <label>更新日期　<input aria-label="更新日期" type="date" value={selected} max={status.data?.today}
        onChange={e => setChosen({page: requestedDate, date: e.target.value})} disabled={busy}/></label>
      <Button type="primary" loading={busy && !supplementing} disabled={!provider || !selected || !status.data || status.isError || busy || selected > status.data.today}
        onClick={() => start.mutate()}>{supplementing ? "核心已更新" : "更新数据"}</Button>
      <Button loading={status.isFetching} disabled={start.isPending}
        onClick={() => void status.refetch()}>查看更新状态</Button>
    </div>
    <small aria-live="polite">{status.isError ? "无法读取更新服务状态" : start.isError ? "未能启动更新，请重试" :
      run?.state ? progress : status.data?.auto ? "交易日收盘后自动更新核心复盘" :
        "更新所选日行情与 AI 建议"}</small>
  </div>;
}

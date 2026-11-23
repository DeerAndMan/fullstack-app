import { useEffect, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button } from "antd";
import { z } from "zod";

import { RequestGet, RequestPost } from "@/api/request-schema";
import styles from "./review.module.css";

const runSchema = z.object({
  date: z.string(), id: z.number(), state: z.enum(["", "running", "failed", "done", "unpublished"]), message: z.string(),
});
const statusSchema = z.object({ today: z.string(), allowed: z.boolean(), windowChangeMs: z.number(), run: runSchema });

function useRefreshStatus() {
  return useQuery({
    queryKey: ["review-refresh-status"],
    queryFn: async ({signal}) => (await RequestGet("/api/v1/review/refresh", {schema: statusSchema, signal})).data,
    refetchInterval: q => {
      const data = q.state.data;
      if (!data) return 15000;
      return Math.max(250, Math.min(data.run.state === "running" ? 2500 : 30000,
        data.windowChangeMs - (Date.now() - q.state.dataUpdatedAt) + 50));
    },
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
    type={run.state === "failed" ? "warning" : run.state === "running" ? "info" : "success"}
    message={run.date + (run.state === "failed" ? " 更新未完成" : run.state === "running" ? " 正在更新" : " 更新完成")}
    description={run.message + "　" + viewing}/>;
}

export default function ReviewUpdate({selectDate}: {selectDate: (date: string) => void}) {
  const client = useQueryClient();
  const [chosen, setChosen] = useState("");
  const watching = useRef<number | undefined>(undefined);
  const status = useRefreshStatus();
  const selected = chosen || status.data?.today || "";
  const start = useMutation({
    mutationFn: async () => (await RequestPost("/api/v1/review/refresh", {date: selected}, {schema: runSchema})).data,
    onSuccess: run => {
      watching.current = run.id;
      client.setQueryData<z.infer<typeof statusSchema>>(["review-refresh-status"], old => old && {...old, run});
      void status.refetch();
    },
  });
  const run = status.data?.run;
  useEffect(() => {
    if (!run) return;
    if (run.state === "running") watching.current = run.id;
    if (run.state !== "done" || watching.current !== run.id) return;
    watching.current = undefined;
    void client.invalidateQueries({predicate: q => typeof q.queryKey[0] === "string" &&
      ["review-dates","review-day","review-timeline","review-study","review-members"].includes(q.queryKey[0])});
    selectDate(run.date);
  }, [run, client, selectDate]);

  const busy = start.isPending || run?.state === "running";
  return <div className={styles.updateBox}>
    <div className={styles.actions}>
      <label>更新日期　<input aria-label="更新日期" type="date" value={selected} max={status.data?.today}
        onChange={e => setChosen(e.target.value)} disabled={busy}/></label>
      <Button type="primary" loading={busy} disabled={!selected || !status.data?.allowed || status.isError || busy || selected > status.data.today}
        onClick={() => start.mutate()}>更新数据</Button>
    </div>
    <small aria-live="polite">{status.isError ? "无法读取更新服务状态" : start.isError ? "未能启动更新，请重试" :
      run?.state === "running" ? run.date + " · " + run.message :
      !status.data?.allowed ? "北京时间 15:00 起可更新" :
      run?.state ? run.date + " · " + run.message : "更新所选日行情、专题和 AI 建议"}</small>
  </div>;
}

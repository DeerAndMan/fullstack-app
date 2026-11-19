import { useEffect, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Skeleton } from "antd";
import { BASE_API } from "@/api";
import { useAuthStore } from "@/stores/auth";
import { readResearchRun, runSchema } from "./research-data";
import type { ResearchRun } from "./research-data";
import ResearchResult from "./result";
import styles from "./research.module.css";

export default function Turn({
  row,
  account,
  edit,
  expanded,
}: {
  row: ResearchRun;
  account?: number;
  edit: (run: ResearchRun) => void;
  expanded: boolean;
}) {
  const client = useQueryClient();
  const [draft, setDraft] = useState("");
  const [progress, setProgress] = useState("正在连接实时更新");
  const detail = useQuery({
    queryKey: ["topic-research", account, "run", row.id],
    queryFn: () => readResearchRun(row.id),
  });
  const run = detail.data ?? row;
  useEffect(() => {
    if (run.status !== "running") return;
    let disposed = false,
      socket: WebSocket,
      retry: ReturnType<typeof setTimeout>;
    const key = ["topic-research", account, "run", row.id];
    function finish(result: ResearchRun) {
      client.setQueryData(key, result);
      void client.invalidateQueries({ queryKey: ["topic-research", account, "runs"] });
    }
    function connect() {
      const url = new URL("/api/v1/topic-research/stream", BASE_API || location.origin);
      url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
      socket = new WebSocket(url);
      socket.onopen = () => socket.send(JSON.stringify({ token: useAuthStore.getState().token, runID: row.id }));
      socket.onmessage = e => {
        if (disposed) return;
        const event = JSON.parse(e.data);
        if (event.type === "reply") setDraft(event.text);
        if (event.type === "progress") setProgress(event.message);
        if (event.type === "result") {
          const parsed = runSchema.safeParse(event.run);
          if (parsed.success) finish(parsed.data);
        }
      };
      socket.onclose = async () => {
        if (disposed) return;
        try {
          const saved = await readResearchRun(row.id);
          if (disposed) return;
          if (saved.status !== "running") {
            finish(saved);
            return;
          }
        } catch {
          /* Retry the authenticated stream after reconnecting. */
        }
        if (!disposed) {
          setProgress("连接中断，正在重连");
          retry = setTimeout(connect, 2000);
        }
      };
    }
    connect();
    return () => {
      disposed = true;
      clearTimeout(retry);
      socket?.close();
    };
  }, [run.status, row.id, account, client]);
  return (
    <article className={styles.turn}>
      <header className={styles.questionBubble}>
        <span>你</span>
        <p>{row.question}</p>
        <button onClick={() => edit(run)}>编辑后重问</button>
      </header>
      {run.status === "running" ? (
        <>
          <p role="status" className={styles.activity}>
            {progress}…
          </p>
          {draft && <p className={styles.reply}>{draft}</p>}
        </>
      ) : detail.isError ? (
        <Alert
          type="error"
          message="读取本轮回答失败"
          action={<Button onClick={() => void detail.refetch()}>重试</Button>}
        />
      ) : !detail.data ? (
        <Skeleton active />
      ) : (
        <>
          {run.result?.reply && <p className={styles.reply}>{run.result.reply}</p>}
          <details open={expanded && !run.result?.reply} className={styles.answer}>
            <summary>{run.result?.reply ? "查看完整研究资料" : "研究回答"}</summary>
            <ResearchResult run={run} />
          </details>
        </>
      )}
    </article>
  );
}

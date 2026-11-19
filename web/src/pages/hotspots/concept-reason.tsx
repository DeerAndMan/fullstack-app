import { Button, Tag } from "antd";

import { useConceptReason } from "./data";
import styles from "./hotspots.module.css";

export function ConceptReason({ concept, symbol, missingMessage = "暂未查到具体关联出处。" }: {
  concept: string;
  symbol: string;
  missingMessage?: string;
}) {
  const query = useConceptReason(concept, symbol);
  const reason = query.data;

  if (query.isPending) return <span className={styles.muted}>正在查询出处…</span>;
  if (query.isError || reason?.status === "unavailable") {
    return <>
      <span className={styles.muted}>资料暂不可用</span>
      <Button type="link" size="small" loading={query.isFetching} onClick={() => void query.refetch()}>重试</Button>
    </>;
  }
  if (reason?.status !== "ready" || !reason.text) return <span className={styles.muted}>{missingMessage}</span>;

  return <>
    {reason.kind && <Tag className={styles.reasonTag}>{reason.kind}</Tag>}
    <span>{reason.text}</span>
    {!!reason.sources?.length && <div className={styles.reasonSources}>
      {reason.sources.map(source => (
        <a key={source.url} href={source.url} target="_blank" rel="noreferrer">{source.title}</a>
      ))}
    </div>}
  </>;
}

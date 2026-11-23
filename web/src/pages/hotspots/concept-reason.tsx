import { Button, Tag } from "antd";

import { TextPreview } from "@/components/text-preview";
import { useConceptReason } from "./data";
import styles from "./hotspots.module.css";

export function ConceptReason({ concept, symbol, missingMessage = "暂未查到具体关联出处。", compact = false }: {
  concept: string;
  symbol: string;
  missingMessage?: string;
  compact?: boolean;
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

  const sources = reason.sources?.map(source => {
    const link = <a key={source.url} href={source.url} target="_blank" rel="noreferrer">{source.title}</a>;
    return compact ? <TextPreview key={source.url} text={source.title}>{link}</TextPreview>
      : link;
  });
  return <div className={compact ? styles.reasonCompact : undefined}>
    {reason.kind && <Tag className={styles.reasonTag}>{reason.kind}</Tag>}
    {compact ? <TextPreview text={reason.text} className={styles.reasonText} /> : <span>{reason.text}</span>}
    {!!sources?.length && <div className={styles.reasonSources}>{sources}</div>}
  </div>;
}

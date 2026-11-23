import { TextPreview } from "@/components/text-preview";
import styles from "./hotspots.module.css";

import type { MarketRow } from "./data";

export function NewsList({ rows, fullDate = false }: { rows: MarketRow[]; fullDate?: boolean }) {
  return (
    <ul className={styles.headlines}>
      {rows.map(row => (
        <li key={row.code}>
          {row.date && <time dateTime={row.date.replace(" ", "T") + (!fullDate && row.date.includes(" ") ? "+08:00" : "")}>
            {fullDate || row.date.length <= 10 ? row.date.slice(0, 16) : row.date.slice(5, 16)}
          </time>}
          <TextPreview text={row.name} detail={row.detail} className={styles.headlineTitle}>
            {row.url && <a href={row.url} target="_blank" rel="noreferrer">{row.name}</a>}
          </TextPreview>
          {row.detail && <TextPreview text={row.detail} className={`${styles.headlineSource} ${styles.muted}`} />}
        </li>
      ))}
    </ul>
  );
}

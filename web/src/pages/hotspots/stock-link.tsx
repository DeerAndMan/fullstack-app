import { generatePath, Link, useLocation } from "react-router-dom";

import ROUTER_PATH from "@/router/section/router-path";
import { isStockSymbol } from "./data";
import styles from "./hotspots.module.css";

import type { MarketRow } from "./data";

export function StockLink({ row, label }: { row: MarketRow; label?: string }) {
  const { search } = useLocation();
  const text = label ?? row.name;
  if (!isStockSymbol(row.symbol)) return text;

  const pathname = generatePath(ROUTER_PATH.hotspotStock, { symbol: row.symbol });
  return <Link className={styles.stockLink} to={{ pathname, search }} state={{ stockName: row.name }}>{text}</Link>;
}

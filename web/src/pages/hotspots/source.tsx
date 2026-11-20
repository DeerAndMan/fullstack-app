import { useSyncExternalStore } from "react";
import { Select } from "antd";
import { useSearchParams } from "react-router-dom";
import type { MarketFeed } from "./data";

type SourceOption = { value: MarketFeed; label: string };
type SourceOptions<T extends MarketFeed> = readonly [{ value: T }, ...{ value: T }[]];

const sessionSources = new Map<string, MarketFeed>();
const sourceListeners = new Set<() => void>();

function subscribeSources(listener: () => void) {
  sourceListeners.add(listener);
  window.addEventListener("storage", listener);
  return () => {
    sourceListeners.delete(listener);
    window.removeEventListener("storage", listener);
  };
}

export function useRememberedMarketSource<T extends MarketFeed>(key: string, options: SourceOptions<T>) {
  const storageKey = `market.source.${key}`;
  const remembered = useSyncExternalStore(subscribeSources, () => {
    const sessionValue = sessionSources.get(storageKey);
    if (sessionValue) return sessionValue;
    try {
      return window.localStorage.getItem(storageKey);
    } catch {
      return null;
    }
  }, () => null);
  const fallback = options[0].value;
  const feed = options.find(option => option.value === remembered)?.value ?? fallback;
  const remember = (value: T) => {
    try {
      window.localStorage.setItem(storageKey, value);
      sessionSources.delete(storageKey);
    } catch {
      sessionSources.set(storageKey, value);
    }
    sourceListeners.forEach(listener => listener());
  };
  return [feed, remember] as const;
}

export const conceptSources = [
  { value: "concept", label: "新浪财经" },
  { value: "concept-eastmoney", label: "东方财富" },
  { value: "concept-ths", label: "同花顺" },
] as const;

export const industrySources = [
  { value: "industry", label: "新浪财经" },
  { value: "industry-eastmoney", label: "东方财富" },
  { value: "industry-ths", label: "同花顺" },
] as const;

export const fundSources = [
  { value: "fund-flow", label: "自动选择" },
  { value: "fund-flow-sina", label: "新浪财经" },
  { value: "fund-flow-eastmoney", label: "东方财富" },
  { value: "fund-flow-ths", label: "同花顺" },
] as const;

export const limitSources = [
  { value: "limit-up", label: "东方财富" },
  { value: "limit-up-ths", label: "同花顺" },
] as const;

export function useMarketSource<T extends MarketFeed>(
  key: string, options: SourceOptions<T>,
) {
  const [params, setParams] = useSearchParams();
  const [remembered, remember] = useRememberedMarketSource(key, options);
  const fallback = options[0].value;
  const feed = options.find(option => option.value === params.get(key))?.value ?? remembered;
  const setFeed = (value: T) => {
    remember(value);
    setParams(previous => {
      const next = new URLSearchParams(previous);
      if (value === fallback) next.delete(key);
      else next.set(key, value);
      return next;
    }, { replace: true });
  };
  return [feed, setFeed] as const;
}

export function SourcePicker<T extends MarketFeed>({ label, value, options, onChange }: {
  label: string;
  value: T;
  options: readonly { value: T; label: string }[];
  onChange: (value: T) => void;
}) {
  return <Select<T, SourceOption>
    aria-label={label}
    size="small"
    value={value}
    options={[...options]}
    onChange={onChange}
    style={{ minWidth: 97 }}
  />;
}

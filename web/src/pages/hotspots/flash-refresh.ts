import { useEffect, useState } from "react";

import type { MarketFeed, MarketQuery } from "./data";

const hour = 60 * 60_000;
const day = 24 * hour;
const preferenceKey = "home.flash.refresh-until";
const minDelay = 30_000;
const maxDelay = 120_000;

export function beijingRefreshTimes(now: number) {
  const midnight = Math.floor((now + 8 * hour) / day) * day - 8 * hour;
  return { cutoff: midnight + 16 * hour, nextDay: midnight + day };
}

export function nextFlashDelay(now: number, cutoff: number, random = Math.random()) {
  const delay = minDelay + Math.floor(random * (maxDelay - minDelay + 1));
  return now + delay < cutoff ? delay : undefined;
}

function storedDeadline(fallback: number) {
  try {
    const raw = window.localStorage.getItem(preferenceKey);
    const deadline = Number(raw);
    return raw !== null && Number.isSafeInteger(deadline) && deadline >= 0 ? deadline : fallback;
  } catch {
    return fallback;
  }
}

export function useFlashRefresh(feed: MarketFeed, query: Pick<MarketQuery, "refetch" | "fetchStatus" | "dataUpdatedAt" | "errorUpdatedAt">) {
  const [now, setNow] = useState(Date.now);
  const { cutoff, nextDay } = beijingRefreshTimes(now);
  const [enabledUntil, setEnabledUntil] = useState(() => storedDeadline(cutoff));
  const [countdown, setCountdown] = useState<number | null>();
  const afterCutoff = now >= cutoff;
  const active = !afterCutoff && enabledUntil === cutoff;
  const nextBoundary = afterCutoff ? nextDay : cutoff;
  const { refetch, fetchStatus, dataUpdatedAt, errorUpdatedAt } = query;

  useEffect(() => {
    try {
      window.localStorage.setItem(preferenceKey, String(enabledUntil));
    } catch {
      // Keep the current setting in memory when storage is unavailable.
    }
  }, [enabledUntil]);

  useEffect(() => {
    const updateClock = () => setNow(Date.now());
    const syncPreference = (event: StorageEvent) => {
      if (event.key === preferenceKey && event.newValue !== null) setEnabledUntil(storedDeadline(cutoff));
    };
    const timer = window.setTimeout(updateClock, Math.max(0, nextBoundary - Date.now()));
    window.addEventListener("focus", updateClock);
    window.addEventListener("storage", syncPreference);
    document.addEventListener("visibilitychange", updateClock);
    return () => {
      window.clearTimeout(timer);
      window.removeEventListener("focus", updateClock);
      window.removeEventListener("storage", syncPreference);
      document.removeEventListener("visibilitychange", updateClock);
    };
  }, [cutoff, nextBoundary]);

  useEffect(() => {
    if (!active || fetchStatus !== "idle") {
      setCountdown(undefined);
      return;
    }
    const scheduledFrom = Date.now();
    const delay = nextFlashDelay(scheduledFrom, cutoff);
    if (delay === undefined) {
      setCountdown(null);
      return;
    }
    const refreshAt = scheduledFrom + delay;
    const updateCountdown = () => setCountdown(Math.max(0, Math.ceil((refreshAt - Date.now()) / 1000)));
    updateCountdown();
    const ticker = window.setInterval(updateCountdown, 1000);
    const refreshTimer = window.setTimeout(() => {
      window.clearInterval(ticker);
      setCountdown(undefined);
      if (Date.now() < cutoff) void refetch({ cancelRefetch: false });
      else setNow(Date.now());
    }, delay);
    return () => {
      window.clearInterval(ticker);
      window.clearTimeout(refreshTimer);
    };
  }, [active, cutoff, dataUpdatedAt, errorUpdatedAt, feed, fetchStatus, refetch]);

  function toggle() {
    const time = Date.now();
    const today = beijingRefreshTimes(time).cutoff;
    setNow(time);
    if (time < today) setEnabledUntil(previous => previous === today ? 0 : today);
  }

  return { active, afterCutoff, countdown, toggle };
}

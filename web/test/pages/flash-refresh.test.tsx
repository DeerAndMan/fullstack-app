// @vitest-environment jsdom

import { act } from "react";
import { createRoot } from "react-dom/client";
import { afterEach, describe, expect, it, vi } from "vitest";

import { beijingRefreshTimes, nextFlashDelay, useFlashRefresh } from "@/pages/hotspots/flash-refresh";

import type { MarketQuery } from "@/pages/hotspots/data";

describe("首页快讯刷新", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    window.localStorage.clear();
    vi.unstubAllGlobals();
  });

  it("随机等待 30 秒至 2 分钟，且不安排跨越 16:00 的请求", () => {
    const now = Date.parse("2026-11-19T15:00:00+08:00");
    const { cutoff, nextDay } = beijingRefreshTimes(now);
    expect(cutoff).toBe(Date.parse("2026-11-19T16:00:00+08:00"));
    expect(nextDay).toBe(Date.parse("2026-11-20T00:00:00+08:00"));
    expect(nextFlashDelay(now, cutoff, 0)).toBe(30_000);
    expect(nextFlashDelay(now, cutoff, 0.5)).toBe(75_000);
    expect(nextFlashDelay(now, cutoff, 0.999999)).toBe(120_000);
    expect(nextFlashDelay(cutoff - 30_000, cutoff, 0)).toBeUndefined();
  });

  it("倒计时对应实际请求，跨过 16:00 后停止并可在次日重开", () => {
    vi.useFakeTimers();
    vi.spyOn(Math, "random").mockReturnValue(0);
    vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
    vi.setSystemTime(new Date("2026-11-19T15:59:10+08:00"));
    const refresh = vi.fn().mockResolvedValue(undefined);
    const host = document.createElement("div");
    const root = createRoot(host);
    const query = { refetch: refresh as MarketQuery["refetch"], fetchStatus: "idle" as const, errorUpdatedAt: 0 };
    function Panel({ updatedAt = 0 }: { updatedAt?: number }) {
      const { active, afterCutoff, countdown, toggle } = useFlashRefresh("flash", { ...query, dataUpdatedAt: updatedAt });
      return <>
        <button disabled={afterCutoff} onClick={toggle}>{active ? "刷新中" : "已停止"}</button>
        <output>{countdown === null ? "无计划" : countdown}</output>
      </>;
    }
    const display = () => host.querySelector("output")?.textContent;

    try {
      act(() => root.render(<Panel />));
      expect(display()).toBe("30");
      act(() => vi.advanceTimersByTime(1000));
      expect(display()).toBe("29");
      act(() => vi.advanceTimersByTime(28_999));
      expect(display()).toBe("1");
      expect(refresh).not.toHaveBeenCalled();
      act(() => vi.advanceTimersByTime(1));
      expect(refresh).toHaveBeenCalledTimes(1);
      expect(display()).toBe("");
      act(() => root.render(<Panel updatedAt={Date.now()} />));
      expect(display()).toBe("无计划");
      act(() => vi.advanceTimersByTime(20_000));
      expect(host.querySelector("button")?.disabled).toBe(true);
      act(() => vi.advanceTimersByTime(8 * 60 * 60_000));
      expect(host.querySelector("button")?.disabled).toBe(false);
      expect(host.textContent).toBe("已停止");
      expect(refresh).toHaveBeenCalledTimes(1);

      act(() => host.querySelector("button")?.click());
      expect(host.querySelector("button")?.textContent).toBe("刷新中");
      expect(display()).toBe("30");
      act(() => vi.advanceTimersByTime(30_000));
      expect(refresh).toHaveBeenCalledTimes(2);
    } finally {
      act(() => root.unmount());
    }
  });
});

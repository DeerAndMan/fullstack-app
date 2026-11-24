// @vitest-environment jsdom

import { focusManager, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { expect, it, vi } from "vitest";
import ReviewUpdate, { ReviewRefreshNotice } from "@/pages/hotspots/review-update";

const { get, post } = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn() }));
vi.mock("@/api/request-schema", () => ({ RequestGet: get, RequestPost: post }));

it("空闲停止轮询，手动更新自动跟进到当天归档后停止轮询", async () => {
  vi.useFakeTimers();
  vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
  const host = document.createElement("div");
  const root = createRoot(host);
  const client = new QueryClient();
  const selectDate = vi.fn();
  focusManager.setFocused(true);
  const day = "2026-11-23";
  const running = { date: day, id: 1, state: "running", message: "正在检查归档缺口" };
  const reply = (state: string) => ({data: {today: day, providers: [{id: "public", label: "腾讯／新浪收盘行情", reason: ""}], run: {...running, state}}});
  get.mockResolvedValueOnce(reply("")).mockResolvedValue(reply("done"));
  post.mockResolvedValue({data: running});
  const click = async (label: string) => {
    const button = Array.from(host.querySelectorAll("button")).find(b => b.textContent === label);
    expect(button?.disabled).toBe(false);
    await act(async () => button!.click());
    await act(async () => vi.advanceTimersByTimeAsync(1));
  };
  try {
    await act(async () => root.render(<QueryClientProvider client={client}>
      <ReviewUpdate selectDate={selectDate}/><ReviewRefreshNotice/>
    </QueryClientProvider>));
    await act(async () => vi.advanceTimersByTimeAsync(180_000));
    expect(get).toHaveBeenCalledTimes(1);
    await click("更新数据");
    await act(async () => vi.advanceTimersByTimeAsync(180_000));
    expect(get).toHaveBeenCalledTimes(2);
    await click("查看更新状态");
    expect(get).toHaveBeenCalledTimes(3);
    expect(selectDate).toHaveBeenCalledWith(day);
    expect(post).toHaveBeenCalledTimes(1);
  } finally {
    act(() => root.unmount());
    client.clear();
    focusManager.setFocused(undefined);
    vi.useRealTimers();
    vi.unstubAllGlobals();
  }
});

it("公开行情自动更新在核心发布时就刷新归档，保留正在查看的历史日期", async () => {
  vi.useFakeTimers();
  vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
  focusManager.setFocused(true);
  get.mockReset();
  const host = document.createElement("div");
  const root = createRoot(host);
  const client = new QueryClient();
  const selectDate = vi.fn();
  const queryKey = ["review-dates"];
  client.setQueryData(queryKey, ["2026-11-20"]);
  const today = "2026-11-23";
  const run = {date: today, id: 2, provider: "public", state: "running", message: "正在读取收盘行情"};
  get.mockResolvedValueOnce({data: {today, auto: true, run}}).mockResolvedValue({
    data: {today, auto: true, run: {...run, corePublished: true, message: "正在补充专题和 AI 建议"}},
  });
  try {
    await act(async () => root.render(<QueryClientProvider client={client}>
      <ReviewUpdate selectDate={selectDate}/><ReviewRefreshNotice displayDate="2026-11-20"/>
    </QueryClientProvider>));
    await act(async () => vi.advanceTimersByTimeAsync(20_100));
    expect(host.textContent).toContain("核心复盘已发布");
    const button = Array.from(host.querySelectorAll("button")).find(b => b.textContent === "核心已更新");
    expect(button?.disabled).toBe(true);
    expect(button?.classList.contains("ant-btn-loading")).toBe(false);
    expect(client.getQueryState(queryKey)?.isInvalidated).toBe(true);
    expect(selectDate).not.toHaveBeenCalled();
  } finally {
    act(() => root.unmount());
    client.clear();
    focusManager.setFocused(undefined);
    vi.useRealTimers();
    vi.unstubAllGlobals();
  }
});

it("打开历史交易日并点击更新时提交该日和历史日线来源", async () => {
  vi.useFakeTimers();
  vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
  const host = document.createElement("div");
  const root = createRoot(host);
  const client = new QueryClient();
  const today = "2026-11-24";
  const requestedDate = "2026-11-23";
  const idle = {date: "", state: "", message: "", id: 0};
  get.mockReset().mockResolvedValue({data: {today, providers: [
    {id: "public", label: "腾讯／新浪收盘行情", reason: ""},
    {id: "baostock", label: "BaoStock 免费日线", reason: ""},
  ], run: idle}});
  post.mockReset().mockResolvedValue({data: {...idle, date: requestedDate, state: "running", id: 1}});
  try {
    await act(async () => root.render(<QueryClientProvider client={client}>
      <ReviewUpdate selectDate={vi.fn()} requestedDate={requestedDate}/>
    </QueryClientProvider>));
    await act(async () => vi.advanceTimersByTimeAsync(1));
    expect(host.querySelector("input")?.value).toBe(requestedDate);
    expect(host.querySelector("select")?.value).toBe("baostock");
    const update = Array.from(host.querySelectorAll("button")).find(button => button.textContent === "更新数据");
    expect(update?.disabled).toBe(false);
    await act(async () => update!.click());
    expect(post).toHaveBeenCalledWith("/api/v1/review/refresh", {date: requestedDate, provider: "baostock"}, expect.anything());
  } finally {
    act(() => root.unmount());
    client.clear();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  }
});

// @vitest-environment jsdom

import { act } from "react";
import { createRoot } from "react-dom/client";
import { MemoryRouter, useLocation, useNavigate } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useMarketSource, useRememberedMarketSource } from "@/pages/hotspots/source";

const regionSources = [{ value: "region-fund-flow" }, { value: "region-fund-flow-ths" }] as const;
const downSources = [{ value: "limit-down" }, { value: "limit-down-xuangu" }] as const;
const flashSources = [{ value: "flash-sina" }, { value: "flash-eastmoney" }] as const;

function RegionAndDown() {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  if (pathname !== "/hotspots") {
    return <button onClick={() => void navigate("/hotspots")}>返回热点</button>;
  }
  return <>
    <PanelSources />
    <button onClick={() => void navigate("/data")}>离开热点</button>
  </>;
}

function PanelSources() {
  const [region, setRegion] = useMarketSource("regionSource", regionSources);
  const [down, setDown] = useMarketSource("downSource", downSources);
  return <>
    <output>{region}/{down}</output>
    <button onClick={() => setRegion("region-fund-flow-ths")}>同花顺</button>
    <button onClick={() => setDown("limit-down-xuangu")}>选股宝</button>
    <button onClick={() => setRegion("region-fund-flow")}>恢复东财</button>
  </>;
}

function FlashSource() {
  const [feed, setFeed] = useRememberedMarketSource("flashSource", flashSources);
  return <button onClick={() => setFeed("flash-eastmoney")}>{feed}</button>;
}

describe("行情来源偏好", () => {
  afterEach(() => {
    window.localStorage.clear();
    vi.unstubAllGlobals();
  });

  it("分开记住区域资金和跌停池，重开无参数的热点页也能恢复", () => {
    vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
    const host = document.createElement("div");
    const root = createRoot(host);
    const render = (path: string) => act(() => root.render(
      <MemoryRouter key={path} initialEntries={[path]}><RegionAndDown /></MemoryRouter>,
    ));
    const click = (text: string) => act(() => {
      [...host.querySelectorAll("button")].find(button => button.textContent === text)?.click();
    });
    const selected = () => host.querySelector("output")?.textContent;

    try {
      render("/hotspots");
      click("同花顺");
      click("选股宝");
      expect(selected()).toBe("region-fund-flow-ths/limit-down-xuangu");
      click("离开热点");
      click("返回热点");
      expect(selected()).toBe("region-fund-flow-ths/limit-down-xuangu");

      render("/hotspots?regionSource=region-fund-flow");
      expect(selected()).toBe("region-fund-flow/limit-down-xuangu");
      render("/hotspots");
      expect(selected()).toBe("region-fund-flow-ths/limit-down-xuangu");
      click("恢复东财");
      expect(window.localStorage.getItem("market.source.regionSource")).toBe("region-fund-flow");
      expect(window.localStorage.getItem("market.source.downSource")).toBe("limit-down-xuangu");
    } finally {
      act(() => root.unmount());
    }
  });

  it("快讯共享来源偏好，并使用有效选项恢复", () => {
    vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
    window.localStorage.setItem("market.source.flashSource", "retired-feed");
    const host = document.createElement("div");
    const root = createRoot(host);
    try {
      act(() => root.render(<><FlashSource /><FlashSource /></>));
      const buttons = () => [...host.querySelectorAll("button")].map(button => button.textContent);
      expect(buttons()).toEqual(["flash-sina", "flash-sina"]);
      act(() => host.querySelector("button")?.click());
      expect(buttons()).toEqual(["flash-eastmoney", "flash-eastmoney"]);
      expect(window.localStorage.getItem("market.source.flashSource")).toBe("flash-eastmoney");
    } finally {
      act(() => root.unmount());
    }
  });
});

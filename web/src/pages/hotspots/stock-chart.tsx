import { useEffect, useMemo, useRef, useState } from "react";
import { theme } from "antd";

import { priceText } from "./data";
import styles from "./hotspots.module.css";

import type { MarketRow } from "./data";

const trendColor = "#3579d4";
const averageColor = "#bc780a";
const minutesPerSession = 242;
const plotLeft = 61;
const plotRight = 18;

function sessionMinute(date: string) {
  const hour = Number(date.slice(11, 13));
  const minute = Number(date.slice(14, 16));
  return hour < 13 ? (hour - 9) * 60 + minute - 30 : 121 + (hour - 13) * 60 + minute;
}

const volumeFormat = new Intl.NumberFormat("zh-CN", { maximumFractionDigits: 0 });
export const volumeText = (value?: number) => value === undefined ? "—" : `${volumeFormat.format(value)} 股`;

export function StockChart({ rows, label, intraday, previousClose }: {
  rows: MarketRow[];
  label: string;
  intraday: boolean;
  previousClose?: number;
}) {
  const { token } = theme.useToken();
  const canvas = useRef<HTMLCanvasElement>(null);
  const [size, setSize] = useState({ width: 0, height: 0, ratio: 1 });
  const [cursor, setCursor] = useState<number>();
  const { days, points, slots, min, max, maxVolume, hasAverage } = useMemo(() => {
    const days = intraday ? [...new Set(rows.map(row => row.date!.slice(0, 10)))] : [];
    const points = rows.map((row, index) => ({
      ...row,
      slot: intraday ? days.indexOf(row.date!.slice(0, 10)) * minutesPerSession + sessionMinute(row.date!) : index,
    }));
    const closes = rows.flatMap(row => [row.low ?? row.price!, row.high ?? row.price!, row.average ?? row.price!]);
    let low = Math.min(...closes);
    let high = Math.max(...closes);
    if (previousClose !== undefined) {
      const radius = Math.max(previousClose - low, high - previousClose);
      low = previousClose - radius;
      high = previousClose + radius;
    }
    const padding = Math.max((high - low) * 0.07, Math.abs(high) * 0.001, 0.01);
    return {
      days, points, slots: intraday ? days.length * minutesPerSession : rows.length,
      min: low - padding, max: high + padding, maxVolume: Math.max(1, ...rows.map(row => row.volume ?? 0)),
      hasAverage: intraday && rows.some(row => row.average !== undefined),
    };
  }, [intraday, previousClose, rows]);
  const selected = points[cursor ?? points.length - 1] ?? points.at(-1)!;
  const plotWidth = Math.max(1, size.width - plotLeft - plotRight);
  const step = plotWidth / slots;
  const x = (slot: number) => plotLeft + (slot + 0.5) * step;

  useEffect(() => {
    const element = canvas.current;
    if (!element) return;
    const observer = new ResizeObserver(() => {
      setSize({ width: element.clientWidth, height: element.clientHeight, ratio: window.devicePixelRatio || 1 });
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const ctx = canvas.current?.getContext("2d");
    if (!ctx || !size.width || !size.height) return;
    const { width, height, ratio } = size;
    const right = width - plotRight;
    const top = 14;
    const bottom = Math.round(height * 0.71);
    const volumeTop = bottom + 22;
    const volumeBottom = height - 25;
    const y = (price: number) => top + (max - price) / (max - min) * (bottom - top);
    const at = (slot: number) => plotLeft + (slot + 0.5) * step;
    const horizontal = (atY: number, color: string) => {
      ctx.strokeStyle = color;
      ctx.beginPath();
      ctx.moveTo(plotLeft, atY);
      ctx.lineTo(right, atY);
      ctx.stroke();
    };
    ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
    ctx.clearRect(0, 0, width, height);
    ctx.font = "11px system-ui, sans-serif";
    ctx.textBaseline = "middle";
    ctx.textAlign = "right";
    ctx.fillStyle = token.colorTextSecondary;
    for (let i = 0; i <= 4; i++) {
      const price = max - (max - min) * i / 4;
      const atY = y(price);
      horizontal(atY, token.colorBorderSecondary);
      ctx.fillText(priceText(price), plotLeft - 8, atY);
    }
    ctx.fillText("成交量", plotLeft - 8, volumeTop + 3);
    horizontal(volumeBottom, token.colorBorderSecondary);
    if (previousClose !== undefined) {
      ctx.setLineDash([4, 4]);
      horizontal(y(previousClose), token.colorTextTertiary);
      ctx.setLineDash([]);
    }

    const drawTrend = (field: "price" | "average", color: string) => {
      ctx.strokeStyle = color;
      ctx.lineWidth = field === "price" ? 1.55 : 1.15;
      ctx.beginPath();
      let lastDay: string | undefined;
      for (const point of points) {
        const value = point[field];
        const day = point.date!.slice(0, 10);
        if (value === undefined) {
          lastDay = undefined;
          continue;
        }
        if (lastDay === day) ctx.lineTo(at(point.slot), y(value));
        else ctx.moveTo(at(point.slot), y(value));
        lastDay = day;
      }
      ctx.stroke();
      ctx.lineWidth = 1;
    };
    if (intraday) {
      if (hasAverage) drawTrend("average", averageColor);
      drawTrend("price", trendColor);
    }
    const barWidth = Math.min(13, Math.max(1, step * 0.67));
    points.forEach((point, index) => {
      const px = at(point.slot);
      const last = points[index - 1];
      const reference = intraday
        ? (last?.date?.slice(0, 10) === point.date?.slice(0, 10) ? last.price! : point.price!)
        : point.open!;
      const color = point.price! > reference ? token.colorError : point.price! < reference ? token.colorSuccess : token.colorTextTertiary;
      ctx.fillStyle = color;
      if (!intraday) {
        ctx.strokeStyle = color;
        ctx.beginPath();
        ctx.moveTo(px, y(point.high!));
        ctx.lineTo(px, y(point.low!));
        ctx.stroke();
        ctx.fillRect(px - barWidth / 2, Math.min(y(point.open!), y(point.price!)), barWidth,
          Math.max(1, Math.abs(y(point.open!) - y(point.price!))));
      }
      if (point.volume) {
        const barHeight = (volumeBottom - volumeTop) * point.volume / maxVolume;
        ctx.globalAlpha = 0.56;
        ctx.fillRect(px - barWidth / 2, volumeBottom - barHeight, barWidth, barHeight);
        ctx.globalAlpha = 1;
      }
    });
    if (cursor !== undefined) {
      const px = at(selected.slot);
      ctx.strokeStyle = token.colorTextTertiary;
      ctx.setLineDash([3, 3]);
      ctx.beginPath();
      ctx.moveTo(px, top);
      ctx.lineTo(px, volumeBottom);
      ctx.stroke();
      ctx.setLineDash([]);
    }
    if (intraday || cursor !== undefined) {
      ctx.fillStyle = intraday ? trendColor : token.colorText;
      ctx.beginPath();
      ctx.arc(at(selected.slot), y(selected.price!), 3, 0, Math.PI * 2);
      ctx.fill();
    }

    const labels: { slot: number; text: string }[] = [];
    if (intraday && days.length === 1) {
      labels.push({ slot: 0, text: "09:30" }, { slot: 120.5, text: "11:30 / 13:00" }, { slot: 241, text: "15:00" });
    } else if (intraday) {
      days.forEach((day, index) => {
        labels.push({ slot: (index + 0.5) * minutesPerSession - 0.5, text: day.slice(5) });
        if (index === 0) return;
        const boundary = at(index * minutesPerSession - 0.5);
        ctx.strokeStyle = token.colorBorderSecondary;
        ctx.beginPath();
        ctx.moveTo(boundary, top);
        ctx.lineTo(boundary, volumeBottom);
        ctx.stroke();
      });
    } else {
      const count = Math.min(points.length, width > 550 ? 5 : 3);
      for (let i = 0; i < count; i++) {
        const index = count === 1 ? 0 : Math.round((points.length - 1) * i / (count - 1));
        labels.push({ slot: index, text: points[index].date! });
      }
    }
    ctx.fillStyle = token.colorTextSecondary;
    ctx.textAlign = "center";
    for (const { slot, text } of labels) {
      const half = ctx.measureText(text).width / 2;
      ctx.fillText(text, Math.max(half + 2, Math.min(width - half - 2, at(slot))), height - 10);
    }
  }, [cursor, days, hasAverage, intraday, max, maxVolume, min, points, previousClose, selected, size, step, token]);

  return <>
    <div className={styles.chartReadout}>
      <strong>{selected.date}</strong>
      {intraday ? <>
        <span><i className={styles.trendSwatch} />价格 <b>{priceText(selected.price)}</b></span>
        {hasAverage && <span><i className={styles.averageSwatch} />均价 <b>{priceText(selected.average)}</b></span>}
      </> : <>
        <span>开 <b>{priceText(selected.open)}</b></span>
        <span>高 <b>{priceText(selected.high)}</b></span>
        <span>低 <b>{priceText(selected.low)}</b></span>
        <span>收 <b>{priceText(selected.price)}</b></span>
      </>}
      <span>成交量 <b>{volumeText(selected.volume)}</b></span>
    </div>
    <canvas ref={canvas} width={Math.round(size.width * size.ratio)} height={Math.round(size.height * size.ratio)}
      className={styles.stockCanvas} tabIndex={0} aria-label={`${label}走势图，使用左右方向键查看报价`}
      onPointerMove={event => {
        const target = event.currentTarget.getBoundingClientRect();
        const px = event.clientX - target.left;
        if (px < plotLeft || px > target.width - plotRight) {
          setCursor(undefined);
          return;
        }
        const nearest = points.reduce((best, point, index) =>
          Math.abs(x(point.slot) - px) < Math.abs(x(points[best].slot) - px) ? index : best, 0);
        setCursor(nearest);
      }}
      onPointerLeave={() => setCursor(undefined)}
      onBlur={() => setCursor(undefined)}
      onKeyDown={event => {
        if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
        event.preventDefault();
        const direction = event.key === "ArrowLeft" ? -1 : 1;
        setCursor(current => Math.max(0, Math.min(points.length - 1, (current ?? points.length - 1) + direction)));
      }}
    />
  </>;
}

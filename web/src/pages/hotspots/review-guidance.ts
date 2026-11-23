type Inputs = { up: number; highs: number; above: number; eligible: number; highEligible: number };

export function reviewGuidance(day: Inputs, delta: number | null) {
  const thin = day.up < 1000;
  const hasWindow = day.highEligible > 0;
  const rising = delta !== null && delta > 0;
  const divergence = thin && hasWindow && rising;
  const previousHighs = delta === null ? null : day.highs - delta;
  const ratio = day.eligible > 0 ? day.above * 100 / day.eligible : null;
  const change = delta === null ? "缺少可比前日" : delta === 0 ? "与前日持平" :
    "较前日" + (delta > 0 ? "增加 " : "减少 ") + Math.abs(delta) + " 家";
  const signals = [];
  if (hasWindow && day.highs < 20) signals.push({
    title: "新高股不足 20 家", reason: "一年新高仅 " + day.highs + " 家。",
    next: "按作者的强势股模式，暂停常规新开仓。等新高家数重新超过 20，再恢复筛选。", metric: "high" as const,
  });
  if (divergence) signals.push({
    title: "普跌中仍有新高股增加", reason: "上涨 " + day.up + " 家，低于 1,000 家，新高" + change + "。",
    next: "加入逆势观察名单。若此前已有连续普跌，再核对板块呼应和个股回踩承接，评估左侧机会。", metric: "high" as const,
  });
  if (hasWindow && previousHighs !== null && previousHighs <= 20 && day.highs > 20) signals.push({
    title: "新高家数重新超过 20", reason: "前一交易日 " + previousHighs + " 家，当日 " + day.highs + " 家。",
    next: "恢复筛选候选。在反复进入前排的行业里找活跃强股，等待个股出现分歧和承接。", metric: "high" as const,
  });
  if (thin && !divergence) signals.push({
    title: "上涨不足 1,000 家", reason: "当日上涨 " + day.up + " 家。" + (hasWindow ? "一年新高" + change + "。" : ""),
    next: "保持观察。核对逆势上涨品种，再看随后交易日是否出现新高家数回升。", metric: "up" as const,
  });
  if (ratio !== null && (ratio < 20 || ratio > 70)) signals.push({
    title: ratio < 20 ? "沪深 300 均线上方成份偏少" : "沪深 300 均线上方成份偏多",
    reason: day.above + " / " + day.eligible + " 家高于 MA50，占比 " + ratio.toFixed(2) + "%，" + (ratio < 20 ? "低于 20%。" : "高于 70%。"),
    next: ratio < 20 ? "进入作者观察指数 ETF 左侧买入机会的区域。继续核对背离与企稳迹象，个股另查板块和形态。" : "进入作者评估指数 ETF 减仓的区域。重点核对指数创新高时，站上均线的成份占比是否下降。", metric: "csi" as const,
  });
  const focus = signals[0] ?? {
    title: "观察行业前排的持续性", reason: "已归档指标未触发普跌、新高收缩或重返、均线占比冷热条件。",
    next: "从逐日表比较容量前三的行业，再到行业表核对其二十日涨幅。", metric: null,
  };
  return {focus, others: signals.slice(1), change, ratio,
    gaps: [!hasWindow && "一年新高无有效样本", delta === null && "前日变化不可比", ratio === null && "MA50 无有效成份"].filter(Boolean).join(" · ")};
}

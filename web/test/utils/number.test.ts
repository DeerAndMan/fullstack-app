import { describe, expect, it } from "vitest";

import {
  divide,
  formatNumber,
  formatPercent,
  minus,
  plus,
  round,
  times,
  toPercent,
} from "@/utils/number";

describe("number utils", () => {
  it("使用十进制精度完成加减乘除", () => {
    expect(plus(0.1, 0.2, "0.3")).toBe(0.6);
    expect(minus(1, 0.1, 0.2)).toBe(0.7);
    expect(times(0.1, 0.2, 10)).toBe(0.2);
    expect(divide(0.3, 0.1)).toBe(3);
  });

  it("将空值视为 0，并安全处理除数为 0", () => {
    expect(plus(null, undefined, "")).toBe(0);
    expect(minus(undefined, 1)).toBe(-1);
    expect(times(null, 10)).toBe(0);
    expect(divide(10, 0)).toBe(0);
    expect(divide(undefined, 2)).toBe(0);
  });

  it("按指定小数位进行精确四舍五入", () => {
    expect(round(1.005, 2)).toBe(1.01);
    expect(round("2.345", 2)).toBe(2.35);
    expect(round(-1.005, 2)).toBe(-1.01);
  });

  it("正确转换并格式化百分比", () => {
    expect(toPercent("0.01234")).toBe(1.23);
    expect(toPercent(0.12555, 1)).toBe(12.6);
    expect(formatPercent(0.0123)).toBe("1.23%");
    expect(formatPercent(null, 1)).toBe("0.0%");
  });

  it("输出固定小数位字符串", () => {
    expect(formatNumber(1.005, 2)).toBe("1.01");
    expect(formatNumber("12", 3)).toBe("12.000");
    expect(formatNumber(undefined)).toBe("0.00");
  });
});

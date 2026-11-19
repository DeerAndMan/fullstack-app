import Decimal from "decimal.js";

/**
 * 数字精度工具
 *
 * 统一封装 decimal.js，用于金额、百分比、涨跌幅等需要精确计算的场景，
 * 避免原生浮点运算（如 0.1 + 0.2 = 0.30000000000000004）和原生 toFixed
 * 的舍入误差。
 *
 * 约定：入参允许 number / string / Decimal / null / undefined，
 * 空值（null / undefined / "" / NaN）统一按 0 处理。
 */

/** 允许的入参类型 */
type Numeric = Decimal.Value | null | undefined;

/** 将任意入参安全转为 Decimal，空值按 0 处理 */
const toDecimal = (value: Numeric): Decimal => new Decimal(value || 0);

/** 加法：plus(a, b, c, ...) 返回各项之和 */
export const plus = (...values: Numeric[]): number =>
  values.reduce<Decimal>((acc, v) => acc.plus(toDecimal(v)), new Decimal(0)).toNumber();

/** 减法：minus(a, b, c) = a - b - c */
export const minus = (value: Numeric, ...subtractors: Numeric[]): number =>
  subtractors.reduce<Decimal>((acc, v) => acc.minus(toDecimal(v)), toDecimal(value)).toNumber();

/** 乘法：times(a, b, c, ...) 返回各项之积 */
export const times = (...values: Numeric[]): number =>
  values.reduce<Decimal>((acc, v) => acc.times(toDecimal(v)), new Decimal(1)).toNumber();

/** 除法：divide(a, b) = a / b，除数为 0 时返回 0 */
export const divide = (value: Numeric, divisor: Numeric): number => {
  const d = toDecimal(divisor);
  return d.isZero() ? 0 : toDecimal(value).div(d).toNumber();
};

/**
 * 精确四舍五入到指定小数位，返回数字。
 * 用于参与后续计算或图表数据的场景。
 */
export const round = (value: Numeric, decimals = 2): number =>
  toDecimal(value).toDecimalPlaces(decimals).toNumber();

/**
 * 转百分比数字：value * 100，四舍五入到指定小数位，返回数字。
 * 例：toPercent(0.0123) => 1.23
 */
export const toPercent = (value: Numeric, decimals = 2): number =>
  toDecimal(value).times(100).toDecimalPlaces(decimals).toNumber();

/**
 * 格式化为固定小数位字符串（decimal.js 版 toFixed，无原生舍入误差）。
 * 例：formatNumber(1.005, 2) => "1.01"
 */
export const formatNumber = (value: Numeric, decimals = 2): string =>
  toDecimal(value).toFixed(decimals);

/**
 * 转百分比字符串：value * 100，固定小数位并带 % 号。
 * 例：formatPercent(0.0123) => "1.23%"
 */
export const formatPercent = (value: Numeric, decimals = 2): string =>
  `${toDecimal(value).times(100).toFixed(decimals)}%`;

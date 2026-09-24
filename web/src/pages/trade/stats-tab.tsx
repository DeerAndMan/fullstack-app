import { useMemo } from "react";
import { Button, Card, Col, DatePicker, Empty, Row, Segmented, Space, Spin, Statistic, Table, Tag } from "antd";
import { DualAxes } from "@ant-design/charts";
import dayjs from "dayjs";

import { groupByOptions, tradeStatsQuery } from "@/api/trade";
import { useTheme } from "@/theme/antd-context";
import { formatNumber, formatPercent, minus, plus, round, times, toPercent } from "@/utils/number";

import type { ColumnsType } from "antd/es/table";
import type { RangePickerProps } from "antd/es/date-picker";
import type { GroupByType } from "@/api/trade";
import type { PeriodStats } from "@/types/schema";

const { RangePicker } = DatePicker;

/**
 * 盈亏色板：正红负绿，与控制台/行情习惯一致。
 * 暗色主题下换用更亮的色阶，否则 #cf1322 / #3f8600 压在黑底上几乎看不清。
 */
const PROFIT_COLORS = {
  light: { up: "#cf1322", down: "#3f8600" },
  dark: { up: "#ff4d4f", down: "#52c41a" },
} as const;

/**
 * 折线色板：三条折线各用一色，避免和盈亏红绿撞色。
 * 当期收益率用蓝、累计盈亏用金、累计收益率用紫，明暗主题下均有足够对比度。
 */
const LINE_COLORS = {
  rate: "#5B8FF9",
  cumProfit: "#FAAD14",
  cumRate: "#9254DE",
} as const;

/** 趋势图上的单个数据点 */
type TrendPoint = {
  period: string;
  区间盈亏: number;
  累计盈亏: number;
  收益率: number;
  累计收益率: number;
};

/** 两条累计线的字段名 */
type CumField = "累计盈亏" | "累计收益率";

/** 单条累计线在当前区间内的极值信息 */
type Extreme = {
  maxIndex: number;
  minIndex: number;
  maxValue: number;
  minValue: number;
  maxPeriod: string;
  minPeriod: string;
};

/**
 * 汇总统计视图：按 周度/月度(按天)/半年度/年度/整体 粒度聚合，
 * 日期区间可自定义，展示整体指标卡片、趋势图与分期明细。
 */
export default function StatsTab() {
  const { statsQuery, rangeQuery, stateOperations, operations } = tradeStatsQuery();
  const { groupBy, setGroupBy, dateRange, setDateRange, bounds } = stateOperations;
  const { theme: mode } = useTheme();

  const isDark = mode === "dark";
  const palette = isDark ? PROFIT_COLORS.dark : PROFIT_COLORS.light;
  // 图表数值标签用中性高对比色，不再跟柱子同色（红字压红柱会糊成一片）
  const labelFill = isDark ? "rgba(255, 255, 255, 0.88)" : "rgba(0, 0, 0, 0.85)";

  const profitColor = (value: number) =>
    value > 0 ? palette.up : value < 0 ? palette.down : undefined;

  const list = statsQuery.data?.data.list ?? [];
  const overall = statsQuery.data?.data.overall ?? null;
  // 日期范围未就绪前也视为加载中，避免闪现「暂无数据」
  const loading = statsQuery.isLoading || rangeQuery.isLoading || !dateRange;

  const pickerChange = (val: RangePickerProps["value"]) => {
    if (val instanceof Array && val[0] && val[1]) {
      setDateRange([dayjs(val[0]), dayjs(val[1])]);
    }
  };

  // 趋势图数据：「区间盈亏」柱 + 三条折线（按日期正序，时间轴从左往右）：
  //   当期收益率、累计盈亏、累计收益率。
  // 两条累计线都从查询区间的起点起算，所以选月度看到的就是本月累计，
  // 周度 / 半年度 / 年度 / 整体同理，跟着查询区间走。
  // 累计盈亏按金额直接相加；累计收益率按日收益率复利连乘，口径与后端 overall 一致。
  // 字段名直接用中文，G2 的图例会拿 yField 当系列名，这样不会显示成 profit / rate。
  const trendData = useMemo<TrendPoint[]>(() => {
    let cumProfit = 0;
    let nav = 1;

    return list.map(p => {
      cumProfit = plus(cumProfit, p.totalProfit);
      nav = times(nav, plus(1, p.profitRate));

      return {
        period: p.periodLabel,
        区间盈亏: round(p.totalProfit),
        累计盈亏: round(cumProfit),
        收益率: toPercent(p.profitRate),
        累计收益率: toPercent(minus(nav, 1)),
      };
    });
  }, [list]);

  // 两条累计线的极值：在当前查询区间内分别取「累计盈亏」「累计收益率」的最高 / 最低点。
  // 同一份结果既用于图上打标，也用于趋势卡片右上角的摘要文字。
  // 粒度或日期区间变化时 trendData 会重算，极值自然跟着当前区间走。
  const trendExtremes = useMemo(() => {
    if (!trendData.length) return null;

    const pick = (field: CumField): Extreme => {
      let maxIndex = 0;
      let minIndex = 0;

      trendData.forEach((d, i) => {
        if (d[field] > trendData[maxIndex][field]) maxIndex = i;
        if (d[field] < trendData[minIndex][field]) minIndex = i;
      });

      return {
        maxIndex,
        minIndex,
        maxValue: trendData[maxIndex][field],
        minValue: trendData[minIndex][field],
        maxPeriod: trendData[maxIndex].period,
        minPeriod: trendData[minIndex].period,
      };
    };

    return { 累计盈亏: pick("累计盈亏"), 累计收益率: pick("累计收益率") };
  }, [trendData]);

  // 明细表反序：最新日期排最上面，和图表的时间轴方向刻意区分开
  const tableData = useMemo(() => [...list].reverse(), [list]);

  // 数据点多时挂上缩略滑块，避免横轴挤在一起
  const needSlider = trendData.length > 24;

  const percentFormatter = (v: number) => `${v}%`;

  /**
   * 累计线的极值标签：只在最高 / 最低点写数值，其余点输出空串。
   * selector 先按数据索引筛掉非极值点；text 再按值兜一层底，
   * 这样即使 selector 因图形结构调整失效，也不会退化成给每个点都打标。
   */
  const extremeLabel = (field: CumField, ext: Extreme, color: string, format: (v: number) => string) => ({
    text: (d: TrendPoint) =>
      d[field] === ext.maxValue
        ? `最高 ${format(d[field])}`
        : d[field] === ext.minValue
          ? `最低 ${format(d[field])}`
          : "",
    selector: (labels: { index: number }[]) =>
      labels.filter(l => l.index === ext.maxIndex || l.index === ext.minIndex),
    style: {
      fontSize: 11,
      fontWeight: 600,
      fill: color,
      textBaseline: "bottom",
      // 最高点标签上抬、最低点压到点下方，避免两个标签都贴着折线互相打架
      dy: (d: TrendPoint) => (d[field] === ext.maxValue ? -6 : 16),
    },
    // exceedAdjust：贴近画布边缘时自动内收，避免标签被裁掉
    transform: [{ type: "exceedAdjust" }],
  });

  const trendConfig = {
    data: trendData,
    xField: "period",
    height: 340,
    legend: true,
    // 图表自身也要跟随明暗主题，否则暗色模式下坐标轴/图例文字仍是深色，几乎看不见
    theme: isDark ? "classicDark" : "classic",
    // 顶部留白，给柱子上方的数值标签腾出空间
    insetTop: 18,
    axis: {
      // 标签自动旋转 + 自动隐藏，标签过密时不再竖排堆叠
      x: { labelAutoRotate: true, labelAutoHide: true, labelAutoEllipsis: true },
    },
    // 图例颜色按 children 顺序指定，保证图例色块和图上实际的柱 / 线颜色一致
    scale: {
      color: {
        range: [palette.up, LINE_COLORS.cumProfit, LINE_COLORS.rate, LINE_COLORS.cumRate],
      },
    },
    ...(needSlider ? { slider: { x: { values: [0.7, 1] } } } : {}),
    children: [
      {
        type: "interval",
        yField: "区间盈亏",
        // 与「累计盈亏」共用同一个金额刻度（左轴），两者单位相同，不能各画各的
        scale: { y: { key: "money" } },
        // 盈利红、亏损绿
        style: { fill: (d: { 区间盈亏: number }) => (d.区间盈亏 >= 0 ? palette.up : palette.down) },
        // 数值标签：正负柱一律画在柱体上边缘之外（负柱即 0 轴上方），
        // 颜色用中性高对比色，不再跟柱子同色，避免红字压红柱看不清
        label: {
          text: (d: { 区间盈亏: number }) => formatNumber(d.区间盈亏),
          position: "top",
          style: {
            fontSize: 11,
            fill: labelFill,
            textBaseline: "bottom",
            dy: -4,
          },
          // exceedAdjust：贴近画布顶部时自动内收，避免标签被裁掉
          transform: [{ type: "exceedAdjust" }],
        },
        axis: { y: { position: "left", title: "盈亏金额" } },
        tooltip: { items: [{ channel: "y", name: "区间盈亏", valueFormatter: formatNumber }] },
      },
      {
        // 累计盈亏：查询区间内逐期累加的盈亏金额，走左侧金额轴
        type: "line",
        yField: "累计盈亏",
        scale: { y: { key: "money" } },
        style: { stroke: LINE_COLORS.cumProfit, lineWidth: 2 },
        // 只在区间内的最高 / 最低点打标，不给每个点都写数值
        ...(trendExtremes
          ? {
              label: extremeLabel(
                "累计盈亏",
                trendExtremes.累计盈亏,
                LINE_COLORS.cumProfit,
                formatNumber,
              ),
            }
          : {}),
        // 不写 axis：共享 key 的各层 axis 配置会被 deepMix 合并，
        // 这里写 false 反而可能把柱子那层的左轴一起关掉，留空即复用同一根轴
        tooltip: { items: [{ channel: "y", name: "累计盈亏", valueFormatter: formatNumber }] },
      },
      {
        // 当期收益率：每一期自身的收益率，走右侧百分比轴
        type: "line",
        yField: "收益率",
        scale: { y: { key: "rate" } },
        style: { stroke: LINE_COLORS.rate, lineWidth: 2 },
        axis: {
          y: {
            position: "right",
            title: "收益率(%)",
            style: { titleFill: LINE_COLORS.rate },
            labelFormatter: percentFormatter,
          },
        },
        tooltip: {
          items: [{ channel: "y", name: "收益率", valueFormatter: percentFormatter }],
        },
      },
      {
        // 累计收益率：区间内按日收益率复利连乘得到，与右侧百分比轴共用刻度
        type: "line",
        yField: "累计收益率",
        scale: { y: { key: "rate" } },
        style: { stroke: LINE_COLORS.cumRate, lineWidth: 2 },
        // 同上：只标区间内的最高 / 最低点
        ...(trendExtremes
          ? {
              label: extremeLabel(
                "累计收益率",
                trendExtremes.累计收益率,
                LINE_COLORS.cumRate,
                percentFormatter,
              ),
            }
          : {}),
        // 同上：右侧百分比轴由「收益率」那一层定义，这里留空复用
        tooltip: {
          items: [{ channel: "y", name: "累计收益率", valueFormatter: percentFormatter }],
        },
      },
    ],
  };

  const columns: ColumnsType<PeriodStats> = [
    { title: "周期", dataIndex: "periodLabel", key: "periodLabel", fixed: "left", width: 130 },
    {
      title: "区间",
      key: "range",
      width: 190,
      render: (_, r) => `${r.startDate} ~ ${r.endDate}`,
    },
    { title: "交易日", dataIndex: "tradingDays", key: "tradingDays", width: 80 },
    {
      title: "区间盈亏",
      dataIndex: "totalProfit",
      key: "totalProfit",
      width: 120,
      render: (v: number) => <span style={{ color: profitColor(v) }}>{formatNumber(v)}</span>,
    },
    {
      title: "收益率",
      dataIndex: "profitRate",
      key: "profitRate",
      width: 100,
      render: (v: number) => <span style={{ color: profitColor(v) }}>{formatPercent(v)}</span>,
    },
    {
      title: "日均盈亏",
      dataIndex: "avgDailyProfit",
      key: "avgDailyProfit",
      width: 110,
      render: (v: number) => <span style={{ color: profitColor(v) }}>{formatNumber(v)}</span>,
    },
    {
      title: "胜率",
      dataIndex: "winRate",
      key: "winRate",
      width: 140,
      render: (v: number, r) => (
        <Space size={4}>
          <span>{formatPercent(v)}</span>
          <Tag color="red">{r.positiveDays}</Tag>
          <Tag color="green">{r.negativeDays}</Tag>
        </Space>
      ),
    },
    {
      title: "最大单日盈利",
      dataIndex: "maxDailyProfit",
      key: "maxDailyProfit",
      width: 120,
      render: (v: number) => <span style={{ color: profitColor(v) }}>{formatNumber(v)}</span>,
    },
    {
      title: "最大单日亏损",
      dataIndex: "minDailyProfit",
      key: "minDailyProfit",
      width: 120,
      render: (v: number) => <span style={{ color: profitColor(v) }}>{formatNumber(v)}</span>,
    },
    {
      title: "最大回撤",
      dataIndex: "maxDrawdown",
      key: "maxDrawdown",
      width: 150,
      render: (v: number, r) => (
        <span>
          {formatPercent(v)}
          {r.maxDrawdownDate ? <span className="text-gray-400 ml-1">({r.maxDrawdownDate})</span> : null}
        </span>
      ),
    },
    {
      title: "年化波动率",
      dataIndex: "volatility",
      key: "volatility",
      width: 110,
      render: (v: number) => formatPercent(v),
    },
    {
      title: "期末资产",
      dataIndex: "endAssets",
      key: "endAssets",
      width: 130,
      render: (v: number) => formatNumber(v),
    },
  ];

  return (
    <>
      <div className="flex justify-between items-center mb-4 flex-wrap gap-3">
        <Space size="middle">
          <span className="text-gray-500">汇总粒度</span>
          <Segmented<GroupByType>
            options={groupByOptions}
            value={groupBy}
            onChange={val => setGroupBy(val)}
          />
        </Space>

        <Space size="middle">
          <span className="text-gray-500">汇总日期</span>
          <RangePicker
            value={dateRange ?? undefined}
            // 可选范围夹在真实数据边界内，避免选到没有快照的空区间
            minDate={bounds?.min}
            maxDate={bounds?.max ?? dayjs()}
            format="YYYY-MM-DD"
            allowClear={false}
            onChange={pickerChange}
          />
          {bounds ? (
            <span className="text-xs text-gray-400">
              数据范围 {bounds.min.format("YYYY-MM-DD")} ~ {bounds.max.format("YYYY-MM-DD")}
            </span>
          ) : null}
          <Button type="primary" onClick={operations.refresh} loading={statsQuery.isFetching}>
            查询
          </Button>
          <Button onClick={() => operations.sync({})} loading={operations.syncing}>
            同步快照
          </Button>
        </Space>
      </div>

      <Spin spinning={loading}>
        {overall ? (
          <>
            {/* 整体汇总指标 */}
            <Row gutter={[16, 16]} className="mb-4">
              <Col xs={12} sm={8} lg={4}>
                <Card size="small">
                  <Statistic
                    title="累计盈亏"
                    value={formatNumber(overall.totalProfit)}
                    valueStyle={{ color: profitColor(overall.totalProfit) }}
                  />
                </Card>
              </Col>
              <Col xs={12} sm={8} lg={4}>
                <Card size="small">
                  <Statistic
                    title="累计收益率"
                    value={formatPercent(overall.profitRate)}
                    valueStyle={{ color: profitColor(overall.profitRate) }}
                  />
                </Card>
              </Col>
              <Col xs={12} sm={8} lg={4}>
                <Card size="small">
                  <Statistic title="胜率" value={toPercent(overall.winRate)} suffix="%" />
                  <div className="text-xs text-gray-400 mt-1">
                    {overall.positiveDays} 盈 / {overall.negativeDays} 亏 / {overall.flatDays} 平
                  </div>
                </Card>
              </Col>
              <Col xs={12} sm={8} lg={4}>
                <Card size="small">
                  <Statistic title="最大回撤" value={toPercent(overall.maxDrawdown)} suffix="%" />
                  <div className="text-xs text-gray-400 mt-1">
                    {overall.maxDrawdownDate || "—"}
                  </div>
                </Card>
              </Col>
              <Col xs={12} sm={8} lg={4}>
                <Card size="small">
                  <Statistic title="年化波动率" value={toPercent(overall.volatility)} suffix="%" />
                </Card>
              </Col>
              <Col xs={12} sm={8} lg={4}>
                <Card size="small">
                  <Statistic title="交易日数" value={overall.tradingDays} suffix="天" />
                  <div className="text-xs text-gray-400 mt-1">
                    日均 {formatNumber(overall.avgDailyProfit)}
                  </div>
                </Card>
              </Col>
            </Row>

            {/* 趋势：盈亏柱（带数值标签）+ 累计盈亏/收益率折线双轴，标题右侧给出两条累计线的极值 */}
            <Card
              size="small"
              title="趋势"
              className="mb-4"
              extra={
                trendExtremes ? (
                  <Space size="large" wrap className="text-xs">
                    <span>
                      <span className="text-gray-400">累计盈亏</span>
                      <span className="ml-2" style={{ color: LINE_COLORS.cumProfit }}>
                        最高 {formatNumber(trendExtremes.累计盈亏.maxValue)}（
                        {trendExtremes.累计盈亏.maxPeriod}）／ 最低{" "}
                        {formatNumber(trendExtremes.累计盈亏.minValue)}（
                        {trendExtremes.累计盈亏.minPeriod}）
                      </span>
                    </span>
                    <span>
                      <span className="text-gray-400">累计收益率</span>
                      <span className="ml-2" style={{ color: LINE_COLORS.cumRate }}>
                        最高 {trendExtremes.累计收益率.maxValue}%（
                        {trendExtremes.累计收益率.maxPeriod}）／ 最低{" "}
                        {trendExtremes.累计收益率.minValue}%（
                        {trendExtremes.累计收益率.minPeriod}）
                      </span>
                    </span>
                  </Space>
                ) : null
              }
            >
              <DualAxes {...trendConfig} />
            </Card>

            {/* 分期明细 */}
            <Table<PeriodStats>
              rowKey="periodLabel"
              size="small"
              columns={columns}
              dataSource={tableData}
              scroll={{ x: 1400 }}
              pagination={{ pageSize: 20, showSizeChanger: true, showTotal: t => `共 ${t} 期` }}
            />
          </>
        ) : (
          !loading && (
            <Empty description="该区间暂无汇总数据，可点击「同步快照」生成日终数据" />
          )
        )}
      </Spin>
    </>
  );
}

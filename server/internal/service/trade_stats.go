package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"

	"gorm.io/gorm"
)

// 支持的汇总粒度
const (
	GroupByDay      = "day"       // 单日
	GroupByWeek     = "week"      // 周度
	GroupByMonth    = "month"     // 月度
	GroupByHalfYear = "half_year" // 半年度
	GroupByYear     = "year"      // 年度
	GroupByCustom   = "custom"    // 自定义区间整体汇总为一组
)

// TradeStatsService 交易统计分析服务。
// 负责从 summary 提取日终快照，并按 日/周/月/半年/年/自定义区间 聚合。
type TradeStatsService struct {
	statsRepo *repository.TradeStatsRepository
}

func NewTradeStatsService(statsRepo *repository.TradeStatsRepository) *TradeStatsService {
	return &TradeStatsService{statsRepo: statsRepo}
}

// ---------- 日终快照同步 ----------

// SyncDailyRequest 日终快照同步请求
type SyncDailyRequest struct {
	StartDate string `json:"startDate"` // 可选，"2006-01-02"；缺省取 2000-01-01（全量回填）
	EndDate   string `json:"endDate"`   // 可选，"2006-01-02"；缺省取今天
}

// SyncDailyResult 同步结果
type SyncDailyResult struct {
	Synced    int    `json:"synced"`    // 本次写入/更新的交易日数
	StartDate string `json:"startDate"` // 实际同步区间
	EndDate   string `json:"endDate"`
}

// SyncDailySnapshots 从 summary 提取日终快照并 upsert 到日终表。
// 盘中重复调用会覆盖当天记录，收盘后最后一次推送即为最终日终值。
func (s *TradeStatsService) SyncDailySnapshots(req *SyncDailyRequest) (*SyncDailyResult, error) {
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate == "" {
		startDate = "2000-01-01"
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return nil, errcode.New(400, fmt.Sprintf("startDate 格式错误，应为 2006-01-02：%v", err), 400)
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		return nil, errcode.New(400, fmt.Sprintf("endDate 格式错误，应为 2006-01-02：%v", err), 400)
	}

	rows, err := s.statsRepo.ExtractDailyFromSummary(startDate+" 00:00:00", endDate+" 23:59:59")
	if err != nil {
		return nil, errcode.New(500, fmt.Sprintf("提取日终快照失败：%v", err), 500)
	}
	if err := s.statsRepo.UpsertDaily(rows); err != nil {
		return nil, errcode.New(500, fmt.Sprintf("写入日终快照失败：%v", err), 500)
	}

	return &SyncDailyResult{Synced: len(rows), StartDate: startDate, EndDate: endDate}, nil
}

// GetDateRange 返回日终快照的可用日期范围，供前端把默认查询区间对齐到真实数据。
func (s *TradeStatsService) GetDateRange() (*repository.DateRange, error) {
	out, err := s.statsRepo.GetDateRange()
	if err != nil {
		return nil, errcode.New(500, fmt.Sprintf("查询日终快照日期范围失败：%v", err), 500)
	}
	return out, nil
}

// ---------- 汇总查询 ----------

// StatsQueryRequest 汇总查询请求。
// groupBy 决定分组粒度，startDate / endDate 为任意自定义闭区间。
type StatsQueryRequest struct {
	GroupBy   string `json:"groupBy" vd:"$=='day' || $=='week' || $=='month' || $=='half_year' || $=='year' || $=='custom'"`
	StartDate string `json:"startDate" vd:"len($)>0"`
	EndDate   string `json:"endDate" vd:"len($)>0"`
}

// PeriodStats 单个周期的汇总指标
type PeriodStats struct {
	PeriodLabel string `json:"periodLabel"` // 展示标签：2025-01-03 / 2025-W01 / 2025-01 / 2025-H1 / 2025
	StartDate   string `json:"startDate"`   // 周期内首个交易日
	EndDate     string `json:"endDate"`     // 周期内最后交易日
	TradingDays int    `json:"tradingDays"` // 交易日数

	// 盈亏（基于每日日终盈亏累加，不受出入金影响）
	TotalProfit    float64 `json:"totalProfit"`    // 区间累计交易盈亏
	ProfitRate     float64 `json:"profitRate"`     // 区间收益率（小数，按日收益率复利连乘）
	AvgDailyProfit float64 `json:"avgDailyProfit"` // 日均盈亏
	MaxDailyProfit float64 `json:"maxDailyProfit"` // 最大单日盈利
	MinDailyProfit float64 `json:"minDailyProfit"` // 最大单日亏损

	// 胜率
	PositiveDays int     `json:"positiveDays"` // 盈利天数
	NegativeDays int     `json:"negativeDays"` // 亏损天数
	FlatDays     int     `json:"flatDays"`     // 持平天数
	WinRate      float64 `json:"winRate"`      // 胜率（小数）

	// 资产（含出入金影响，仅作参考）
	StartAssets  float64 `json:"startAssets"`  // 期初总资产
	EndAssets    float64 `json:"endAssets"`    // 期末总资产
	MaxAssets    float64 `json:"maxAssets"`    // 期间最高总资产
	MinAssets    float64 `json:"minAssets"`    // 期间最低总资产
	AssetsChange float64 `json:"assetsChange"` // 资产变动 = 期末 - 期初（= 交易盈亏 + 净出入金）

	// 风险（基于日收益率构建的净值曲线，剔除出入金影响）
	MaxDrawdown     float64 `json:"maxDrawdown"`     // 最大回撤（小数）
	MaxDrawdownDate string  `json:"maxDrawdownDate"` // 最大回撤发生日
	Volatility      float64 `json:"volatility"`      // 年化波动率（252 交易日）
}

// StatsQueryResponse 汇总查询响应
type StatsQueryResponse struct {
	GroupBy string        `json:"groupBy"`
	List    []PeriodStats `json:"list"`    // 按粒度分组的趋势序列
	Overall *PeriodStats  `json:"overall"` // 整个查询区间的总汇总
}

// QueryStats 按粒度 + 自定义区间查询汇总统计。
func (s *TradeStatsService) QueryStats(req *StatsQueryRequest) (*StatsQueryResponse, error) {
	if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
		return nil, errcode.New(400, fmt.Sprintf("startDate 格式错误，应为 2006-01-02：%v", err), 400)
	}
	if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
		return nil, errcode.New(400, fmt.Sprintf("endDate 格式错误，应为 2006-01-02：%v", err), 400)
	}
	if req.StartDate > req.EndDate {
		return nil, errcode.New(400, "startDate 不能晚于 endDate", 400)
	}

	daily, err := s.statsRepo.ListDailyByDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, errcode.New(500, fmt.Sprintf("查询日终快照失败：%v", err), 500)
	}
	resp := &StatsQueryResponse{GroupBy: req.GroupBy, List: []PeriodStats{}}
	if len(daily) == 0 {
		return resp, nil
	}

	// 区间首日的期初资产取「首个交易日之前最近一条」的期末资产；
	// 没有更早数据时，用首日资产回推（期末 - 当日盈亏）。
	baseAssets := daily[0].Zzc - daily[0].Dryk
	if prev, err := s.statsRepo.GetLastDailyBefore(daily[0].TradeDate); err == nil {
		baseAssets = prev.Zzc
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(500, fmt.Sprintf("查询期初资产失败：%v", err), 500)
	}

	// 分组并逐组计算；组间期初资产用上一组期末衔接，避免 N+1 查询。
	groups := groupDaily(daily, req.GroupBy)
	carry := baseAssets
	for _, g := range groups {
		ps := buildPeriodStats(g, req.GroupBy, carry)
		resp.List = append(resp.List, ps)
		carry = ps.EndAssets
	}

	// 整体汇总始终按 custom 口径计算一次
	overall := buildPeriodStats(daily, GroupByCustom, baseAssets)
	resp.Overall = &overall

	return resp, nil
}

// ---------- 内部计算 ----------

// groupDaily 按粒度把日终快照切分成若干组（组内按交易日升序，组间按时间升序）。
func groupDaily(daily []model.TradeDailySummary, groupBy string) [][]model.TradeDailySummary {
	// custom 整个区间视为一组
	if groupBy == GroupByCustom {
		return [][]model.TradeDailySummary{daily}
	}
	// day 每个交易日一组
	if groupBy == GroupByDay {
		out := make([][]model.TradeDailySummary, 0, len(daily))
		for i := range daily {
			out = append(out, daily[i:i+1])
		}
		return out
	}

	bucket := make(map[string][]model.TradeDailySummary)
	for _, row := range daily {
		key := periodKey(row.TradeDate, groupBy)
		bucket[key] = append(bucket[key], row)
	}

	keys := make([]string, 0, len(bucket))
	for k := range bucket {
		keys = append(keys, k)
	}
	sort.Strings(keys) // 各粒度 key 均为零填充定长格式，字典序即时间序

	out := make([][]model.TradeDailySummary, 0, len(keys))
	for _, k := range keys {
		out = append(out, bucket[k])
	}
	return out
}

// periodKey 生成分组键，格式保证零填充定长以便字典序排序。
func periodKey(tradeDate, groupBy string) string {
	t, err := time.Parse("2006-01-02", tradeDate)
	if err != nil {
		return tradeDate
	}
	switch groupBy {
	case GroupByWeek:
		// 用 ISO 周，跨年周归属以 ISO 标准为准
		year, week := t.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case GroupByMonth:
		return t.Format("2006-01")
	case GroupByHalfYear:
		half := "H1"
		if t.Month() >= 7 {
			half = "H2"
		}
		return fmt.Sprintf("%04d-%s", t.Year(), half)
	case GroupByYear:
		return fmt.Sprintf("%04d", t.Year())
	default:
		return tradeDate
	}
}

// buildPeriodStats 计算一个周期的全部指标。
// baseAssets 为该周期的期初总资产（上一交易日期末值）。
func buildPeriodStats(group []model.TradeDailySummary, groupBy string, baseAssets float64) PeriodStats {
	first, last := group[0], group[len(group)-1]

	ps := PeriodStats{
		StartDate:      first.TradeDate,
		EndDate:        last.TradeDate,
		TradingDays:    len(group),
		StartAssets:    baseAssets,
		EndAssets:      last.Zzc,
		MaxDailyProfit: first.Dryk,
		MinDailyProfit: first.Dryk,
		MaxAssets:      first.Zzc,
		MinAssets:      first.Zzc,
	}

	switch groupBy {
	case GroupByCustom:
		ps.PeriodLabel = fmt.Sprintf("%s ~ %s", first.TradeDate, last.TradeDate)
	case GroupByDay:
		ps.PeriodLabel = first.TradeDate
	default:
		ps.PeriodLabel = periodKey(first.TradeDate, groupBy)
	}

	// 净值曲线：由每日收益率复利连乘得到，天然剔除出入金影响
	navCurve := make([]float64, 0, len(group))
	nav := 1.0

	var sumProfit float64
	for _, row := range group {
		sumProfit += row.Dryk

		switch {
		case row.Dryk > 0:
			ps.PositiveDays++
		case row.Dryk < 0:
			ps.NegativeDays++
		default:
			ps.FlatDays++
		}

		if row.Dryk > ps.MaxDailyProfit {
			ps.MaxDailyProfit = row.Dryk
		}
		if row.Dryk < ps.MinDailyProfit {
			ps.MinDailyProfit = row.Dryk
		}
		if row.Zzc > ps.MaxAssets {
			ps.MaxAssets = row.Zzc
		}
		if row.Zzc < ps.MinAssets {
			ps.MinAssets = row.Zzc
		}

		nav *= 1 + dailyReturn(row)
		navCurve = append(navCurve, nav)
	}

	ps.TotalProfit = sumProfit
	ps.AvgDailyProfit = sumProfit / float64(ps.TradingDays)
	ps.AssetsChange = ps.EndAssets - ps.StartAssets
	ps.ProfitRate = nav - 1
	if ps.TradingDays > 0 {
		ps.WinRate = float64(ps.PositiveDays) / float64(ps.TradingDays)
	}
	ps.MaxDrawdown, ps.MaxDrawdownDate = maxDrawdown(group, navCurve)
	ps.Volatility = annualizedVolatility(group)

	return ps
}

// maxDailyReturn 单日收益率的合理上限（绝对值）。
// 超过 100% 基本可以断定是脏数据（历史上 summary 曾把盈亏金额写进 drhz 列），
// 直接采信会让净值连乘瞬间放大成天文数字，必须改走兜底算法。
const maxDailyReturn = 1.0

// dailyReturn 取当日收益率（小数）。
// 优先用 summary 侧算好的 drhz；缺失或量级明显异常时，
// 退回「当日盈亏 / 当日期初资产」重算。
func dailyReturn(row model.TradeDailySummary) float64 {
	if row.Drhz != 0 && math.Abs(row.Drhz) < maxDailyReturn {
		return row.Drhz
	}
	openAssets := row.Zzc - row.Dryk
	if openAssets > 0 {
		r := row.Dryk / openAssets
		// 兜底结果同样做量级校验，避免期初资产残缺时算出畸形收益率
		if math.Abs(r) < maxDailyReturn {
			return r
		}
	}
	return 0
}

// maxDrawdown 基于净值曲线计算最大回撤及其发生日期。
func maxDrawdown(group []model.TradeDailySummary, navCurve []float64) (float64, string) {
	var maxDD float64
	var ddDate string
	// 期初净值恒为 1，以它作为初始峰值，否则首日即亏损时这段回撤会被漏掉
	peak := 1.0

	for i, nav := range navCurve {
		if nav > peak {
			peak = nav
		}
		if peak > 0 {
			if dd := (peak - nav) / peak; dd > maxDD {
				maxDD = dd
				ddDate = group[i].TradeDate
			}
		}
	}
	return maxDD, ddDate
}

// annualizedVolatility 日收益率标准差年化（按 252 个交易日）。
func annualizedVolatility(group []model.TradeDailySummary) float64 {
	if len(group) < 2 {
		return 0
	}

	returns := make([]float64, 0, len(group))
	for _, row := range group {
		returns = append(returns, dailyReturn(row))
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	var variance float64
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	// 样本标准差，自由度 n-1
	variance /= float64(len(returns) - 1)

	return math.Sqrt(variance) * math.Sqrt(252)
}

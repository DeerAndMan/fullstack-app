package repository

import (
	"strings"

	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TradeStatsRepository 交易统计数据访问层。
// 负责「summary 原始推送 -> 日终快照」的提取，以及日终快照表的读写。
type TradeStatsRepository struct {
	db *gorm.DB
}

func NewTradeStatsRepository(db *gorm.DB) *TradeStatsRepository {
	return &TradeStatsRepository{db: db}
}

// extractDailySQL 从 summary 表提取每个交易日的日终快照。
//
// 关键点：
//  1. 子查询按 DATE(date) 分组，用 MAX(id) 锁定当天最后一条推送（id 自增，写入顺序即时间顺序）；
//  2. 同一子查询顺带算出当天推送条数与日内盈亏极值；
//  3. summary 中数值以字符串存储，统一 CAST 成 DECIMAL；NULLIF(TRIM(x),”) 规避空串，
//     外层再包 COALESCE 兜底，避免空串/NULL 写入 NOT NULL 列时报 1048；
//  4. trade_date 用 DATE_FORMAT 输出成 'YYYY-MM-DD' 字符串，避免 DSN 的 parseTime=True
//     把 DATE 列扫描成 time.Time 再转成 RFC3339（会导致上层按日期分组失效）。
//
// 历史数据存在两种推送格式，必须分别适配，否则字段语义错位：
//   - 老格式（约 2025-03 之前）：dryk / zzc 为 NULL，drhz 里存的其实是「当日盈亏金额」，
//     总资产落在 zsz 字段；
//   - 新格式：dryk 为当日盈亏金额、zzc 为期末总资产，drhz 才是真正的当日收益率（小数）。
//
// 因此以「dryk 是否有值」判别格式：老格式把 drhz 归位到 dryk、zsz 归位到 zzc，
// 并把 drhz 置 0，交由 service 用「当日盈亏 / 当日期初资产」重算收益率。
const extractDailySQL = "SELECT\n" +
	"  DATE_FORMAT(s.`date`, '%Y-%m-%d')               AS trade_date,\n" +
	// 当日盈亏：新格式取 dryk；老格式 dryk 缺失，drhz 存的才是盈亏金额
	"  COALESCE(\n" +
	"    CAST(NULLIF(TRIM(s.dryk), '') AS DECIMAL(20,4)),\n" +
	"    CAST(NULLIF(TRIM(s.drhz), '') AS DECIMAL(20,4)),\n" +
	"    0\n" +
	"  )                                               AS dryk,\n" +
	// 当日收益率：仅新格式的 drhz 是比例，老格式置 0 由 service 兜底重算
	"  CASE WHEN NULLIF(TRIM(s.dryk), '') IS NULL THEN 0\n" +
	"       ELSE COALESCE(CAST(NULLIF(TRIM(s.drhz), '') AS DECIMAL(16,8)), 0)\n" +
	"  END                                             AS drhz,\n" +
	// 期末总资产：新格式取 zzc；老格式无 zzc，用 zsz（总市值）兜底
	"  COALESCE(\n" +
	"    CAST(NULLIF(TRIM(s.zzc), '') AS DECIMAL(20,4)),\n" +
	"    CAST(NULLIF(TRIM(s.zsz), '') AS DECIMAL(20,4)),\n" +
	"    0\n" +
	"  )                                               AS zzc,\n" +
	"  COALESCE(CAST(NULLIF(TRIM(s.zxsz), '') AS DECIMAL(20,4)), 0) AS zxsz,\n" +
	"  COALESCE(CAST(NULLIF(TRIM(s.zjye), '') AS DECIMAL(20,4)), 0) AS zjye,\n" +
	"  COALESCE(s.ljyk, 0)                             AS ljyk,\n" +
	"  COALESCE(s.kyzj, 0)                             AS kyzj,\n" +
	"  COALESCE(s.num, 0)                              AS position_count,\n" +
	"  d.cnt                                           AS snapshot_count,\n" +
	"  COALESCE(d.max_dryk, 0)                         AS intraday_max_dryk,\n" +
	"  COALESCE(d.min_dryk, 0)                         AS intraday_min_dryk,\n" +
	"  s.id                                            AS last_summary_id\n" +
	"FROM summary s\n" +
	"JOIN (\n" +
	"  SELECT DATE(`date`) AS d,\n" +
	"         MAX(id)      AS mid,\n" +
	"         COUNT(*)     AS cnt,\n" +
	// 日内极值同样要兼容两种格式，统一取「盈亏金额」所在的那一列
	"         MAX(COALESCE(CAST(NULLIF(TRIM(dryk), '') AS DECIMAL(20,4)),\n" +
	"                      CAST(NULLIF(TRIM(drhz), '') AS DECIMAL(20,4)))) AS max_dryk,\n" +
	"         MIN(COALESCE(CAST(NULLIF(TRIM(dryk), '') AS DECIMAL(20,4)),\n" +
	"                      CAST(NULLIF(TRIM(drhz), '') AS DECIMAL(20,4)))) AS min_dryk\n" +
	"  FROM summary\n" +
	"  WHERE `date` >= ? AND `date` <= ?\n" +
	"  GROUP BY DATE(`date`)\n" +
	") d ON s.id = d.mid\n" +
	"ORDER BY trade_date ASC"

// ExtractDailyFromSummary 按时间区间从 summary 提取日终快照（不落库）。
// startTime / endTime 为 "2006-01-02 15:04:05" 格式。
func (r *TradeStatsRepository) ExtractDailyFromSummary(startTime, endTime string) ([]model.TradeDailySummary, error) {
	var rows []model.TradeDailySummary
	err := r.db.Raw(extractDailySQL, startTime, endTime).Scan(&rows).Error
	normalizeTradeDates(rows)
	return rows, err
}

// UpsertDaily 按 trade_date 唯一键写入或更新日终快照。
// 盘中重复推送时会不断覆盖当天记录，收盘后的最后一次即为最终日终值。
func (r *TradeStatsRepository) UpsertDaily(rows []model.TradeDailySummary) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "trade_date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"dryk", "drhz", "zzc", "zxsz", "zjye", "ljyk", "kyzj",
			"intraday_max_dryk", "intraday_min_dryk",
			"position_count", "snapshot_count", "last_summary_id", "updated_at",
		}),
	}).CreateInBatches(rows, 200).Error
}

// ListDailyByDateRange 读取区间内的日终快照，按交易日升序。
// startDate / endDate 为 "2006-01-02" 格式，闭区间。
func (r *TradeStatsRepository) ListDailyByDateRange(startDate, endDate string) ([]model.TradeDailySummary, error) {
	var list []model.TradeDailySummary
	err := r.db.Where("trade_date >= ? AND trade_date <= ?", startDate, endDate).
		Order("trade_date ASC").
		Find(&list).Error
	normalizeTradeDates(list)
	return list, err
}

// GetLastDailyBefore 取指定交易日之前最近的一条日终快照，用于计算区间首日的期初资产。
func (r *TradeStatsRepository) GetLastDailyBefore(date string) (*model.TradeDailySummary, error) {
	var row model.TradeDailySummary
	err := r.db.Where("trade_date < ?", date).Order("trade_date DESC").First(&row).Error
	if err != nil {
		return nil, err
	}
	row.TradeDate = normalizeTradeDate(row.TradeDate)
	return &row, nil
}

// DateRange 日终快照的可用日期范围。
type DateRange struct {
	MinDate string `json:"minDate"` // 最早交易日，无数据时为空串
	MaxDate string `json:"maxDate"` // 最晚交易日，无数据时为空串
	Days    int64  `json:"days"`    // 交易日总数
}

// GetDateRange 返回日终快照表的日期范围，供前端设置默认查询区间。
// 用 DATE_FORMAT 直接输出字符串，规避 NULL 与时间戳格式问题。
func (r *TradeStatsRepository) GetDateRange() (*DateRange, error) {
	var out DateRange
	err := r.db.Model(&model.TradeDailySummary{}).
		Select("COALESCE(DATE_FORMAT(MIN(trade_date), '%Y-%m-%d'), '') AS min_date, " +
			"COALESCE(DATE_FORMAT(MAX(trade_date), '%Y-%m-%d'), '') AS max_date, " +
			"COUNT(*) AS days").
		Scan(&out).Error
	return &out, err
}

// CountDaily 返回日终快照总行数。
func (r *TradeStatsRepository) CountDaily() (int64, error) {
	var total int64
	err := r.db.Model(&model.TradeDailySummary{}).Count(&total).Error
	return total, err
}

// normalizeTradeDate 把交易日统一裁剪成 "2006-01-02"。
// DSN 开启 parseTime=True 时，MySQL 的 DATE 列会被扫描成 time.Time 再格式化成
// RFC3339（如 "2024-09-24T00:00:00+08:00"），上层按日期解析分组会直接失败。
func normalizeTradeDate(date string) string {
	if len(date) > 10 {
		return date[:10]
	}
	return strings.TrimSpace(date)
}

// normalizeTradeDates 批量规范化交易日字段，原地修改。
func normalizeTradeDates(rows []model.TradeDailySummary) {
	for i := range rows {
		rows[i].TradeDate = normalizeTradeDate(rows[i].TradeDate)
	}
}

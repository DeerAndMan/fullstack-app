package model

import "time"

// TradeDailySummary 交易日终快照表。
//
// 背景：summary 表由油猴脚本在盘中持续推送写入，同一交易日会产生几十条记录，
// 且 dryk（当日盈亏）是「日内累计值」而非增量，直接对 summary 做 SUM 会把同一天重复累加。
// 因此这里按交易日固化一行「日终快照」（取当天最后一条推送），
// 周 / 月 / 半年 / 年以及任意自定义区间的汇总都在本表上聚合。
//
// 数值字段在 summary 中以字符串存储，落到本表时统一转成 decimal，便于索引与聚合。
type TradeDailySummary struct {
	ID        uint   `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	TradeDate string `json:"tradeDate" gorm:"column:trade_date;type:date;not null;uniqueIndex:uk_trade_date;comment:交易日"`

	// 日终值（取当天最后一条推送）
	Dryk float64 `json:"dryk" gorm:"column:dryk;type:decimal(20,4);not null;default:0;comment:当日盈亏(日终值)"`
	Drhz float64 `json:"drhz" gorm:"column:drhz;type:decimal(16,8);not null;default:0;comment:当日盈亏比例(小数,非百分数)"`
	Zzc  float64 `json:"zzc" gorm:"column:zzc;type:decimal(20,4);not null;default:0;comment:期末总资产"`
	Zxsz float64 `json:"zxsz" gorm:"column:zxsz;type:decimal(20,4);not null;default:0;comment:期末最新市值"`
	Zjye float64 `json:"zjye" gorm:"column:zjye;type:decimal(20,4);not null;default:0;comment:期末资金余额"`
	Ljyk float64 `json:"ljyk" gorm:"column:ljyk;type:decimal(20,4);not null;default:0;comment:累计盈亏"`
	Kyzj float64 `json:"kyzj" gorm:"column:kyzj;type:decimal(20,4);not null;default:0;comment:可用资金"`

	// 日内统计（来自当天全部推送记录）
	IntradayMaxDryk float64 `json:"intradayMaxDryk" gorm:"column:intraday_max_dryk;type:decimal(20,4);not null;default:0;comment:日内最高盈亏"`
	IntradayMinDryk float64 `json:"intradayMinDryk" gorm:"column:intraday_min_dryk;type:decimal(20,4);not null;default:0;comment:日内最低盈亏"`

	PositionCount int  `json:"positionCount" gorm:"column:position_count;not null;default:0;comment:期末持仓只数"`
	SnapshotCount int  `json:"snapshotCount" gorm:"column:snapshot_count;not null;default:0;comment:当日推送快照条数"`
	LastSummaryID uint `json:"lastSummaryId" gorm:"column:last_summary_id;not null;default:0;comment:日终快照来源 summary.id"`

	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (TradeDailySummary) TableName() string {
	return "trade_daily_summary"
}

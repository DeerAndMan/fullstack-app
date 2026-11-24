package service

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"

	"github.com/fatih/color"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type EnergyService struct {
	energyRepo *repository.EnergyRepository
	statsSvc   *TradeStatsService
}

func NewEnergyService(energyRepo *repository.EnergyRepository, statsSvc *TradeStatsService) *EnergyService {
	return &EnergyService{energyRepo: energyRepo, statsSvc: statsSvc}
}

func (s *EnergyService) InsertAssets(data []model.Assets) error {
	if len(data) == 0 {
		return nil
	}
	now := time.Now()
	snapshotTime := now.Format("2006-01-02 15:04:05")
	snapshots := make([]repository.AssetSnapshot, 0, len(data))
	for _, item := range data {
		if len(item.Positions) <= 0 {
			return errcode.New(400, "Positions 数据不存在", 400)
		}
		snapshot := repository.AssetSnapshot{Positions: slices.Clone(item.Positions)}
		for i := range snapshot.Positions {
			if snapshot.Positions[i].DATETIME == "" {
				snapshot.Positions[i].DATETIME = snapshotTime
			}
		}
		if err := copier.Copy(&snapshot.Summary, &item); err != nil {
			return errcode.New(500, "构建资产快照失败", 500)
		}
		snapshot.Summary.Date = snapshotTime
		snapshot.Summary.Drhz = "0"
		dryk, drykErr := strconv.ParseFloat(item.Dryk, 64)
		zzc, zzcErr := strconv.ParseFloat(item.Zzc, 64)
		if drykErr == nil && zzcErr == nil && zzc != 0 {
			snapshot.Summary.Drhz = fmt.Sprintf("%.4f", dryk/zzc)
		}
		snapshots = append(snapshots, snapshot)
	}

	if err := s.energyRepo.CreateSnapshots(snapshots); err != nil {
		return errcode.New(500, "保存资产快照失败", 500)
	}

	// 刷新当天日终快照，使汇总统计实时跟随最新推送。
	// 快照同步失败不影响本次数据写入，仅记录告警日志。
	if s.statsSvc != nil {
		today := now.Format("2006-01-02")
		if _, err := s.statsSvc.SyncDailySnapshots(&SyncDailyRequest{StartDate: today, EndDate: today}); err != nil {
			zap.L().Warn("刷新日终快照失败", zap.Error(err))
		}
	}

	// 彩色打印持仓汇总到控制台
	printAssetsInfo(data)

	return nil
}

// colorizeValue 根据数值正负着色：上涨使用红色，下跌使用绿色，零值及无法解析的值不着色。
func colorizeValue(value string) string {
	trimmed := strings.TrimSpace(value)
	parsed := strings.TrimSuffix(trimmed, "%")
	if number, err := strconv.ParseFloat(parsed, 64); err == nil {
		switch {
		case number > 0:
			return color.RedString("%s", value)
		case number < 0:
			return color.GreenString("%s", value)
		}
	}
	return value
}

// colorizeAs 用 reference 数值的正负来给 label 着色。
func colorizeAs(label, reference string) string {
	trimmed := strings.TrimSpace(reference)
	parsed := strings.TrimSuffix(trimmed, "%")
	if number, err := strconv.ParseFloat(parsed, 64); err == nil {
		switch {
		case number > 0:
			return color.RedString("%s", label)
		case number < 0:
			return color.GreenString("%s", label)
		}
	}
	return label
}

// formatPercentValue 将小数形式的比例转换为带百分号的百分比字符串。
// 例如：0.0112 -> 1.12%，-0.0042 -> -0.42%。
func formatPercentValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasSuffix(trimmed, "%") {
		return trimmed
	}

	number, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return value
	}
	return fmt.Sprintf("%.2f%%", number*100)
}

// dailyRatio 返回当日盈亏占总资产的比例，与写入 Summary 的计算规则保持一致。
func dailyRatio(item model.Assets) string {
	dryk, drykErr := strconv.ParseFloat(strings.TrimSpace(item.Dryk), 64)
	zzc, zzcErr := strconv.ParseFloat(strings.TrimSpace(item.Zzc), 64)
	if drykErr == nil && zzcErr == nil && zzc != 0 {
		return fmt.Sprintf("%.4f", dryk/zzc)
	}
	return item.Drhz
}

// printAssetsInfo 彩色打印持仓汇总信息到控制台。
func printAssetsInfo(data []model.Assets) {
	for _, item := range data {
		fmt.Printf("当前日期：%s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("当日盈亏 ---> %s 比例 ---> %s\n",
			colorizeValue(item.Dryk),
			colorizeValue(formatPercentValue(dailyRatio(item))),
		)

		parts := make([]string, 0, len(item.Positions))
		for _, position := range item.Positions {
			parts = append(parts, fmt.Sprintf("%s -> %s %s",
				colorizeAs(position.Zqmc, position.Dryk),
				colorizeValue(position.Dryk),
				colorizeValue(formatPercentValue(position.Drykbl)),
			))
		}
		fmt.Println(strings.Join(parts, " | "))
	}
}

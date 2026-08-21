package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"

	"github.com/fatih/color"
	"github.com/jinzhu/copier"
)

type EnergyService struct {
	energyRepo *repository.EnergyRepository
}

func NewEnergyService(energyRepo *repository.EnergyRepository) *EnergyService {
	return &EnergyService{energyRepo: energyRepo}
}

func (s *EnergyService) InsertAssets(data []model.Assets) error {
	for _, item := range data {
		if len(item.Positions) <= 0 {
			return errcode.New(400, "Positions 数据不存在", 400)
		}
		for i := range item.Positions {
			if item.Positions[i].DATETIME == "" {
				item.Positions[i].DATETIME = time.Now().Format("2006-01-02 15:04:05")
			}
		}
		if err := s.energyRepo.BatchInsertPositions(item.Positions, 100); err != nil {
			return errcode.New(500, "插入 Positions 失败", 500)
		}
	}

	if err := s.insertSummaryData(data); err != nil {
		return err
	}

	// 彩色打印持仓汇总到控制台
	printAssetsInfo(data)

	return nil
}

func (s *EnergyService) insertSummaryData(data []model.Assets) error {
	for _, v := range data {
		summary := model.Summary{}
		if err := copier.Copy(&summary, &v); err != nil {
			return errcode.New(500, "插入 Summary 失败", 500)
		}

		drykNum, drykErr := strconv.ParseFloat(v.Dryk, 64)
		zzcNum, zzcErr := strconv.ParseFloat(v.Zzc, 64)
		if drykErr == nil && zzcErr == nil && zzcNum != 0 {
			summary.Drhz = fmt.Sprintf("%.4f", drykNum/zzcNum)
		} else {
			summary.Drhz = "0"
		}

		summary.Date = time.Now().Format("2006-01-02 15:04:05")
		positionsJson, err := json.Marshal(v.Positions)
		if err != nil {
			return errcode.New(500, "序列化 Positions 失败", 500)
		}
		summary.Positions = string(positionsJson)

		if err := s.energyRepo.CreateSummary(&summary); err != nil {
			return errcode.New(500, "插入 Summary 失败", 500)
		}
	}
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

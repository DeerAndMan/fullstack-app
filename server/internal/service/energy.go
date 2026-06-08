package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"

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

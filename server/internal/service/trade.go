package service

import (
	"encoding/json"
	"fmt"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"
)

type TradeService struct {
	energyRepo *repository.EnergyRepository
}

func NewTradeService(energyRepo *repository.EnergyRepository) *TradeService {
	return &TradeService{energyRepo: energyRepo}
}

type TradeRequest struct {
	StartTime string `json:"startTime" vd:"len($)>0"`
	EndTime   string `json:"endTime" vd:"len($)>0"`
}

type TradeResponse struct {
	List []map[string]interface{} `json:"list"`
}

func (s *TradeService) parseTimeRange(req *TradeRequest) (string, string, error) {
	start, err := time.Parse("2006-01-02", req.StartTime)
	if err != nil {
		return "", "", fmt.Errorf("开始时间解析错误：%v", err)
	}
	end, err := time.Parse("2006-01-02", req.EndTime)
	if err != nil {
		return "", "", fmt.Errorf("结束时间解析错误：%v", err)
	}
	startTime := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()).Format("2006-01-02 15:04:05")
	endTime := time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location()).Format("2006-01-02 15:04:05")
	return startTime, endTime, nil
}

func (s *TradeService) Index(req *TradeRequest) ([]map[string]interface{}, error) {
	startTime, endTime, err := s.parseTimeRange(req)
	if err != nil {
		return nil, errcode.New(400, err.Error(), 400)
	}

	list, err := s.energyRepo.ListSummaryByDateRange(startTime, endTime)
	if err != nil {
		return nil, errcode.New(500, fmt.Sprintf("查询数据失败：%v", err), 500)
	}

	var responseList []map[string]interface{}
	for _, l := range list {
		var dataMap map[string]interface{}
		jsonData, _ := json.Marshal(l)
		json.Unmarshal(jsonData, &dataMap)

		if positionStr, ok := dataMap["positions"].(string); ok && positionStr != "" {
			var positionData interface{}
			if err := json.Unmarshal([]byte(positionStr), &positionData); err == nil {
				dataMap["positions"] = positionData
			}
		}
		responseList = append(responseList, dataMap)
	}

	return responseList, nil
}

type SummaryResponse struct {
	model.Summary
	Positions []model.Energy `json:"positions"`
}

func (s *TradeService) Summary(req *TradeRequest) (*SummaryResponse, error) {
	startTime, endTime, err := s.parseTimeRange(req)
	if err != nil {
		return nil, errcode.New(400, err.Error(), 400)
	}

	summary, err := s.energyRepo.GetSummaryByDateRange(startTime, endTime)
	if err != nil {
		return nil, errcode.New(500, fmt.Sprintf("查询数据失败：%v", err), 500)
	}

	resp := &SummaryResponse{Summary: *summary}
	if summary.Positions != "" {
		var positions []model.Energy
		if err := json.Unmarshal([]byte(summary.Positions), &positions); err != nil {
			return nil, errcode.New(500, fmt.Sprintf("解析 Positions 失败：%v", err), 500)
		}
		resp.Positions = positions
	}

	return resp, nil
}

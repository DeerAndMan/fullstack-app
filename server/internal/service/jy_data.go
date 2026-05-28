package service

import (
	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
)

type JyDataService struct {
	jyDataRepo *repository.JyDataRepository
}

func NewJyDataService(jyDataRepo *repository.JyDataRepository) *JyDataService {
	return &JyDataService{jyDataRepo: jyDataRepo}
}

type JyDataListRequest struct {
	StartTime string `json:"startTime" vd:"len($)>0"`
	EndTime   string `json:"endTime,omitempty"`
}

func (s *JyDataService) GetLatest() ([]model.JyData, error) {
	return s.jyDataRepo.GetLatestByMaxDate()
}

func (s *JyDataService) ListByDateRange(req *JyDataListRequest) ([]model.JyData, error) {
	return s.jyDataRepo.ListByDateRange(req.StartTime, req.EndTime)
}

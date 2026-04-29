package repository

import (
	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
)

type EnergyRepository struct {
	db *gorm.DB
}

func NewEnergyRepository(db *gorm.DB) *EnergyRepository {
	return &EnergyRepository{db: db}
}

func (r *EnergyRepository) BatchInsertPositions(positions []model.Energy, batchSize int) error {
	return r.db.CreateInBatches(positions, batchSize).Error
}

func (r *EnergyRepository) CreateSummary(summary *model.Summary) error {
	return r.db.Create(summary).Error
}

func (r *EnergyRepository) GetLastSummary() (*model.Summary, error) {
	var summary model.Summary
	err := r.db.Omit("positions").Last(&summary).Error
	return &summary, err
}

func (r *EnergyRepository) ListSummaryByDateRange(startTime, endTime string) ([]model.Summary, error) {
	var list []model.Summary
	err := r.db.Where("date >= ? AND date < ?", startTime, endTime).Find(&list).Error
	return list, err
}

func (r *EnergyRepository) GetSummaryByDateRange(startTime, endTime string) (*model.Summary, error) {
	var summary model.Summary
	err := r.db.Where("date >= ? AND date <= ?", startTime, endTime).Find(&summary).Error
	return &summary, err
}

package repository

import (
	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
)

type JyDataRepository struct {
	db *gorm.DB
}

func NewJyDataRepository(db *gorm.DB) *JyDataRepository {
	return &JyDataRepository{db: db}
}

func (r *JyDataRepository) GetLatestByMaxDate() ([]model.JyData, error) {
	var maxDate string
	if err := r.db.Model(&model.JyData{}).Select("MAX(DATE(NOEDATE))").Scan(&maxDate).Error; err != nil {
		return nil, err
	}

	var list []model.JyData
	err := r.db.Where("DATE(NOEDATE) = ?", maxDate).Find(&list).Error
	return list, err
}

func (r *JyDataRepository) ListByDateRange(startTime string, endTime string) ([]model.JyData, error) {
	var list []model.JyData
	query := r.db.Model(&model.JyData{}).
		Where(`STR_TO_DATE(NOEDATE, '%Y-%m-%d %H:%i:%s') >= STR_TO_DATE(?, '%Y-%m-%d %H:%i:%s')`, startTime)

	if endTime != "" {
		query = query.Where(`STR_TO_DATE(NOEDATE, '%Y-%m-%d %H:%i:%s') <= STR_TO_DATE(?, '%Y-%m-%d %H:%i:%s')`, endTime)
	}

	err := query.Find(&list).Error
	return list, err
}

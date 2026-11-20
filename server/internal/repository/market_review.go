package repository

import (
	"context"
	"errors"

	"fullstack-app/server/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MarketReviewRepository struct{ db *gorm.DB }

func NewMarketReviewRepository(db *gorm.DB) *MarketReviewRepository {
	return &MarketReviewRepository{db}
}

func (r *MarketReviewRepository) Publish(ctx context.Context, day *model.MarketReviewDay, members []model.MarketReviewMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Concurrent imports converge on the same revision without exposing a partial list.
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(day).Error; err != nil {
			return err
		}
		var stored model.MarketReviewDay
		if err := tx.Where("batch = ?", day.Batch).Take(&stored).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.MarketReviewMember{}).Where("day_id = ?", stored.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == int64(len(members)) {
			return nil
		}
		if count != 0 {
			return errors.New("incomplete existing review revision")
		}
		for i := range members {
			members[i].DayID = stored.ID
		}
		return tx.CreateInBatches(members, 200).Error
	})
}

func (r *MarketReviewRepository) Dates(ctx context.Context) (dates []string, err error) {
	dates = []string{}
	err = r.db.WithContext(ctx).Model(&model.MarketReviewDay{}).Distinct("date").Order("date DESC").Pluck("date", &dates).Error
	return
}

func (r *MarketReviewRepository) Day(ctx context.Context, date, batch string) (day model.MarketReviewDay, err error) {
	q := r.db.WithContext(ctx).Order("date DESC, id DESC")
	if date != "" {
		q = q.Where("date = ?", date)
	}
	if batch != "" {
		q = q.Where("batch = ?", batch)
	}
	err = q.Take(&day).Error
	return
}

func (r *MarketReviewRepository) Members(ctx context.Context, id uint64, metric, search string, page int) ([]model.MarketReviewMember, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.MarketReviewMember{}).Where("day_id = ?", id)
	switch metric {
	case "up":
		q = q.Where("`change` > 0")
	case "high":
		q = q.Where("new_high = ?", true)
	case "csi":
		q = q.Where("csi = ?", true)
	case "excluded":
		q = q.Where("reason <> ''")
	case "all":
	default:
		return nil, 0, errors.New("unknown metric")
	}
	if search != "" {
		q = q.Where("LOCATE(?,symbol)>0 OR LOCATE(?,name)>0", search, search)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	members := []model.MarketReviewMember{}
	err := q.Order("symbol ASC").Limit(50).Offset((page - 1) * 50).Find(&members).Error
	return members, count, err
}

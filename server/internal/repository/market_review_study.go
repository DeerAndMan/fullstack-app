package repository

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"fullstack-app/server/internal/model"
	"gorm.io/gorm/clause"
)

type ReviewCoverage struct {
	Above        int    `json:"above"`
	Eligible     int    `json:"eligible"`
	PrevAbove    *int   `json:"prevAbove"`
	PrevEligible *int   `json:"prevEligible"`
	Date         string `json:"date"`
	Batch        string `json:"batch"`
	Kinds        string `json:"kinds"`
}

func (r *MarketReviewRepository) StudyCoverage(ctx context.Context) ([]ReviewCoverage, error) {
	var rows []ReviewCoverage
	err := r.db.WithContext(ctx).Raw(`SELECT d.date,d.batch,d.above,d.eligible,p.above AS prev_above,p.eligible AS prev_eligible,COALESCE(GROUP_CONCAT(s.kind),'') AS kinds FROM market_review_days d
	 JOIN (SELECT MAX(id) id FROM market_review_days GROUP BY date) latest ON latest.id=d.id
	 LEFT JOIN market_review_days p ON p.id=(SELECT MAX(prev.id) FROM market_review_days prev WHERE prev.date=d.previous AND prev.rule=d.rule)
	 LEFT JOIN market_review_studies s ON s.day_id=d.id GROUP BY d.id,d.date,d.batch,d.above,d.eligible,p.above,p.eligible ORDER BY d.date`).Scan(&rows).Error
	return rows, err
}

func (r *MarketReviewRepository) SaveStudy(ctx context.Context, date, batch, kind string, payload json.RawMessage) error {
	if kind != "patterns" && kind != "sectors" && kind != "style" {
		return fmt.Errorf("unknown study kind %s", kind)
	}
	if len(payload) == 0 || !json.Valid(payload) || payload[0] != '{' {
		return fmt.Errorf("invalid study payload")
	}
	day, err := r.Day(ctx, date, batch)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	out := gzip.NewWriter(&buf)
	if _, err = out.Write(payload); err != nil {
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	row := model.MarketReviewStudy{DayID: day.ID, Kind: kind, Data: buf.Bytes()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "day_id"}, {Name: "kind"}}, DoUpdates: clause.AssignmentColumns([]string{"data"})}).Create(&row).Error
}

func (r *MarketReviewRepository) Study(ctx context.Context, id uint64, kind string) (json.RawMessage, error) {
	var row model.MarketReviewStudy
	if err := r.db.WithContext(ctx).Where("day_id = ? AND kind = ?", id, kind).Take(&row).Error; err != nil {
		return nil, err
	}
	in, err := gzip.NewReader(bytes.NewReader(row.Data))
	if err != nil {
		return nil, err
	}
	defer in.Close()
	return io.ReadAll(in)
}

func (r *MarketReviewRepository) Note(ctx context.Context, user uint, date, symbol string) (model.MarketReviewNote, error) {
	var note model.MarketReviewNote
	err := r.db.WithContext(ctx).Where("user_id=? AND date=? AND symbol=?", user, date, symbol).Take(&note).Error
	return note, err
}

func (r *MarketReviewRepository) SaveNote(ctx context.Context, note *model.MarketReviewNote) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "date"}, {Name: "symbol"}}, DoUpdates: clause.AssignmentColumns([]string{"text", "updated_at"})}).Create(note).Error
}

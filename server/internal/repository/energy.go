package repository

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
)

type EnergyRepository struct {
	db *gorm.DB
}

type AssetSnapshot struct {
	Summary   model.Summary
	Positions []model.Energy
}

type positionReferences struct {
	EnergyIDs []uint `json:"energyIds"`
}

func NewEnergyRepository(db *gorm.DB) *EnergyRepository {
	return &EnergyRepository{db: db}
}

func (r *EnergyRepository) CreateSnapshots(snapshots []AssetSnapshot) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, snapshot := range snapshots {
			if err := tx.CreateInBatches(snapshot.Positions, 100).Error; err != nil {
				return err
			}
			references := positionReferences{EnergyIDs: make([]uint, len(snapshot.Positions))}
			for i, position := range snapshot.Positions {
				if position.ID == 0 {
					return errors.New("saved position is missing its ID")
				}
				references.EnergyIDs[i] = position.ID
			}
			raw, err := json.Marshal(references)
			if err != nil {
				return err
			}
			snapshot.Summary.Positions = string(raw)
			if err := tx.Create(&snapshot.Summary).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *EnergyRepository) GetLastSummary() (*model.Summary, error) {
	var summary model.Summary
	err := r.db.Omit("positions").Last(&summary).Error
	return &summary, err
}

func (r *EnergyRepository) ListSummaryByDateRange(startTime, endTime string) ([]model.Summary, error) {
	var list []model.Summary
	err := r.db.Where("date >= ? AND date < ?", startTime, endTime).Find(&list).Error
	if err == nil {
		err = r.loadSummaryPositions(list)
	}
	return list, err
}

func (r *EnergyRepository) GetSummaryByDateRange(startTime, endTime string) (*model.Summary, error) {
	summaries := make([]model.Summary, 1)
	err := r.db.Where("date >= ? AND date <= ?", startTime, endTime).Find(&summaries[0]).Error
	if err == nil {
		err = r.loadSummaryPositions(summaries)
	}
	return &summaries[0], err
}

// Expand compact references for callers that expect a complete JSON array.
// Rows written before the change already contain that array.
func (r *EnergyRepository) loadSummaryPositions(summaries []model.Summary) error {
	bySummary := make(map[int][]uint)
	positions := make(map[uint]model.Energy)
	var ids []uint
	for i, summary := range summaries {
		raw := strings.TrimSpace(summary.Positions)
		if !strings.HasPrefix(raw, `{"energyIds":`) {
			continue
		}

		var references positionReferences
		if err := json.Unmarshal([]byte(raw), &references); err != nil {
			return err
		}
		bySummary[i] = references.EnergyIDs
		for _, id := range references.EnergyIDs {
			if _, seen := positions[id]; !seen {
				ids = append(ids, id)
				positions[id] = model.Energy{}
			}
		}
	}
	for batch := range slices.Chunk(ids, 1000) {
		var rows []model.Energy
		if err := r.db.Where("id IN ?", batch).Find(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			positions[row.ID] = row
		}
	}
	for i, references := range bySummary {
		rows := make([]model.Energy, 0, len(references))
		for _, id := range references {
			item := positions[id]
			if item.ID == 0 {
				return errors.New("asset snapshot refers to a missing position")
			}
			rows = append(rows, item)
		}
		raw, err := json.Marshal(rows)
		if err != nil {
			return err
		}
		summaries[i].Positions = string(raw)
	}
	return nil
}

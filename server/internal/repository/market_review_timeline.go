package repository

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fullstack-app/server/internal/model"
	"sort"
)

type ReviewTimelineRow struct {
	Batch           string              `json:"batch"`
	Advice          *model.ReviewAdvice `json:"advice,omitempty"`
	AdvicePending   bool                `json:"advicePending,omitempty"`
	Date            string              `json:"date"`
	Up              int                 `json:"up"`
	Highs           int                 `json:"highs"`
	Above           int                 `json:"above"`
	Eligible        int                 `json:"eligible"`
	BeijingHighs    *int                `json:"beijingHighs"`
	BeijingEligible int                 `json:"beijingEligible,omitempty"`
	HighDelta       *int                `json:"highDelta"`
	AverageState    string              `json:"averageState,omitempty"`
	StyleBias       string              `json:"styleBias,omitempty"`
	AverageChange   *float64            `json:"averageChange"`
	IndexDivergence *bool               `json:"indexDivergence"`
	MarketAmount    *float64            `json:"marketAmount"`
	CapacityLeaders []string            `json:"capacityLeaders"`
}

// Timeline reads just the two small studies for each published day revision.
// Pattern evidence and stock lists stay behind the single-day drilldowns.
func (r *MarketReviewRepository) Timeline(ctx context.Context, through string) ([]ReviewTimelineRow, error) {
	var records []struct {
		Date                              string
		Up, Highs, Above, Eligible        int
		HighDelta                         *int
		Batch                             string
		AdviceData, StyleData, SectorData []byte
	}
	err := r.db.WithContext(ctx).Raw(`SELECT d.batch, a.data AS advice_data, d.date, d.up, d.highs, d.above, d.eligible,
		CASE WHEN p.rule=d.rule THEN d.highs-p.highs END AS high_delta,
		s.data AS style_data, b.data AS sector_data
		FROM (SELECT MAX(id) AS id FROM market_review_days WHERE date<=? GROUP BY date ORDER BY date DESC LIMIT 15) recent
		JOIN market_review_days d ON d.id=recent.id
		LEFT JOIN market_review_days p ON p.id=(SELECT MAX(prev.id) FROM market_review_days prev WHERE prev.date=d.previous)
		LEFT JOIN market_review_studies a ON a.day_id=d.id AND a.kind=?
		LEFT JOIN market_review_studies s ON s.day_id=d.id AND s.kind='style'
		LEFT JOIN market_review_studies b ON b.day_id=d.id AND b.kind='sectors'
		ORDER BY d.date`, through, model.ReviewAdviceKind).Scan(&records).Error
	if err != nil {
		return nil, err
	}
	rows := make([]ReviewTimelineRow, 0, len(records))
	for _, record := range records {
		row := ReviewTimelineRow{Batch: record.Batch, Date: record.Date, Up: record.Up, Highs: record.Highs, Above: record.Above, Eligible: record.Eligible, HighDelta: record.HighDelta}
		if err := row.readStudies(record.StyleData, record.SectorData); err != nil {
			return nil, err
		}
		if record.AdviceData != nil {
			if err := decodeTimelineStudy(record.AdviceData, &row.Advice); err != nil {
				return nil, err
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func decodeTimelineStudy(raw []byte, value any) error {
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer reader.Close()
	return json.NewDecoder(reader).Decode(value)
}

func (r *ReviewTimelineRow) readStudies(style, sectors []byte) error {
	if style != nil {
		var s model.ReviewStyleData
		if err := decodeTimelineStudy(style, &s); err != nil {
			return err
		}
		r.IndexDivergence = s.Index.Divergence
		r.MarketAmount = s.MarketAmount
		if s.Beijing != nil && s.Beijing.Date == r.Date {
			r.BeijingHighs = s.Beijing.Highs
			r.BeijingEligible = s.Beijing.Eligible
		}
		r.AverageState, r.StyleBias = s.Labels()
		r.AdvicePending = s.DeferredAdvice && sectors == nil
		for _, index := range s.Rows {
			if index.Code == "880003" {
				r.AverageChange = index.Change
				break
			}
		}
	}
	if sectors != nil {
		var s struct {
			MarketAmount *float64
			Rows         []struct {
				Code, Name string
				Hits       []string
			}
		}
		if err := decodeTimelineStudy(sectors, &s); err != nil {
			return err
		}
		if r.MarketAmount == nil {
			r.MarketAmount = s.MarketAmount
		}
		sort.Slice(s.Rows, func(i, j int) bool {
			a, b := s.Rows[i], s.Rows[j]
			if len(a.Hits) == len(b.Hits) {
				return a.Code < b.Code
			}
			return len(a.Hits) > len(b.Hits)
		})
		r.CapacityLeaders = []string{}
		for _, sector := range s.Rows[:min(3, len(s.Rows))] {
			if len(sector.Hits) > 0 {
				r.CapacityLeaders = append(r.CapacityLeaders, sector.Name)
			}
		}
	}
	return nil
}

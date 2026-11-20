package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

type ReviewPattern struct {
	Kind        string  `json:"kind"`
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Anchor      string  `json:"anchor"`
	SignalDate  string  `json:"signalDate"`
	TriggerDate string  `json:"triggerDate,omitempty"`
	Days        int     `json:"days"`
	Big         float64 `json:"big"`
	Near        float64 `json:"near"`
	Width       float64 `json:"width"`
	Doji        float64 `json:"doji"`
	Shrink      float64 `json:"shrink"`
	Close       float64 `json:"close"`
	Open        float64 `json:"open"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	EMA5        float64 `json:"ema5"`
	EMA13       float64 `json:"ema13"`
	Amount      float64 `json:"amount"`
}

type ReviewThresholds struct{ Big, Near, Width, Doji, Shrink float64 }

func DefaultReviewThresholds() ReviewThresholds { return ReviewThresholds{5, 2, 8, 35, .75} }

func (t ReviewThresholds) Valid() bool {
	values := []struct{ v, lo, hi float64 }{{t.Big, 1, 30}, {t.Near, 0, 25}, {t.Width, 0, 50}, {t.Doji, 0, 100}, {t.Shrink, 0, 3}}
	for _, v := range values {
		if math.IsNaN(v.v) || math.IsInf(v.v, 0) || v.v < v.lo || v.v > v.hi {
			return false
		}
	}
	return true
}

func FilterReviewPatterns(data json.RawMessage, t ReviewThresholds, kind, code, stage string) ([]ReviewPattern, error) {
	if !t.Valid() {
		return nil, fmt.Errorf("invalid thresholds")
	}
	var input struct{ Rows []ReviewPattern }
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}
	out := []ReviewPattern{}
	for _, r := range input.Rows {
		if (kind != "" && kind != r.Kind) || (code != "" && !strings.Contains(r.Symbol, code) && !strings.Contains(r.Name, code)) || r.Big < t.Big || (stage == "trigger") != (r.TriggerDate != "") {
			continue
		}
		match := false
		switch r.Kind {
		case "上涨中继":
			match = r.Doji <= t.Doji && r.Shrink <= t.Shrink
		case "EMA5 小平台", "EMA5 大平台":
			match = r.Near <= t.Near && r.Width <= t.Width
		case "EMA13 回踩", "超跌回踩":
			match = r.Near <= t.Near
		}
		if match {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Big != out[j].Big {
			return out[i].Big > out[j].Big
		}
		if out[i].Symbol != out[j].Symbol {
			return out[i].Symbol < out[j].Symbol
		}
		return out[i].Kind < out[j].Kind
	})
	return out, nil
}

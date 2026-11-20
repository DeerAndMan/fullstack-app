package model

// Each revision is immutable. Batch is the SHA-256 of the calculation artifact.
type MarketReviewDay struct {
	ID           uint64 `gorm:"primaryKey" json:"-"`
	Date         string `gorm:"size:10;index" json:"date"`
	Previous     string `gorm:"size:10" json:"previous"`
	Rule         string `gorm:"size:64" json:"rule"`
	Batch        string `gorm:"size:64;uniqueIndex" json:"batch"`
	Release      string `gorm:"size:10" json:"release"`
	Source       string `gorm:"size:200" json:"source"`
	WindowStart  string `gorm:"size:10" json:"windowStart"`
	Total        int    `json:"total"`
	Traded       int    `json:"traded"`
	Up           int    `json:"up"`
	HighEligible int    `json:"highEligible"`
	Highs        int    `json:"highs"`
	CSI          int    `json:"csi"`
	Eligible     int    `json:"eligible"`
	Above        int    `json:"above"`
}

type MarketReviewMember struct {
	DayID        uint64   `gorm:"primaryKey;autoIncrement:false" json:"-"`
	Symbol       string   `gorm:"size:12;primaryKey" json:"symbol"`
	Name         string   `gorm:"size:80" json:"name"`
	CSI          bool     `json:"csi"`
	Bars         int      `json:"bars"`
	Change       *float64 `json:"change,omitempty"`
	Close        *float64 `json:"close,omitempty"`
	High         *float64 `json:"high,omitempty"`
	Factor       *float64 `json:"factor,omitempty"`
	PreviousHigh *float64 `json:"previousHigh,omitempty"`
	PeakDate     string   `gorm:"size:10" json:"peakDate,omitempty"`
	NewHigh      *bool    `json:"newHigh,omitempty"`
	MA50         *float64 `json:"ma50,omitempty"`
	Above        *bool    `json:"above,omitempty"`
	Reason       string   `gorm:"size:100" json:"reason,omitempty"`
}

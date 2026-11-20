package model

import "time"

type MarketReviewStudy struct {
	DayID uint64 `gorm:"primaryKey;autoIncrement:false"`
	Kind  string `gorm:"primaryKey;size:16"`
	Data  []byte `gorm:"type:longblob;not null"`
}

type MarketReviewNote struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"-"`
	Date      string    `gorm:"primaryKey;size:10" json:"date"`
	Symbol    string    `gorm:"primaryKey;size:12" json:"symbol"`
	Text      string    `gorm:"size:1000" json:"text"`
	UpdatedAt time.Time `json:"updatedAt"`
}

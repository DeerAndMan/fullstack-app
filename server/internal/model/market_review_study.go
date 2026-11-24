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

// ReviewStyleData is the numerical part of the archived style study.
type ReviewStyleData struct {
	Rows []struct {
		Code                       string
		Change, Close, EMA5, EMA13 *float64
	}
	Spread, MarketAmount *float64
	Beijing              *struct {
		Date     string
		Highs    *int
		Eligible int
	}
	DeferredAdvice bool
	Index          struct{ Divergence *bool }
}

func (data ReviewStyleData) Labels() (string, string) {
	average, bias := "", ""
	if data.Spread != nil {
		switch {
		case *data.Spread > 0:
			bias = "新高强于平均股价"
		case *data.Spread < 0:
			bias = "新高弱于平均股价"
		default:
			bias = "两者涨跌幅相同"
		}
	}
	for _, r := range data.Rows {
		if r.Code != "880003" || r.Close == nil || r.EMA5 == nil || r.EMA13 == nil {
			continue
		}
		position := func(level float64) string {
			switch {
			case *r.Close > level:
				return "高于"
			case *r.Close < level:
				return "低于"
			default:
				return "等于"
			}
		}
		average = position(*r.EMA5) + "EMA5、" + position(*r.EMA13) + "EMA13"
	}
	return average, bias
}

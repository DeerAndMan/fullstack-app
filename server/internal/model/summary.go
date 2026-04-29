package model

type Summary struct {
	ID            int    `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	Date          string `gorm:"type:datetime;not null" json:"date"`
	Drhz          string `json:"drhz"`
	Dryk          string `json:"dryk"`
	Zxsz          string `json:"zxsz"`
	Zzc           string `json:"zzc"`
	RMBZzc        string `gorm:"column:RMBZzc" json:"RMBZzc"`
	Num           int    `gorm:"not null" json:"num"`
	Zsz           string `json:"zsz"`
	Ccyk          string `json:"ccyk"`
	Stocks        string `json:"stocks"`
	Zjye          string `json:"zjye"`
	Positions     string `json:"positions" gorm:"column:positions"`
	Djzj          int    `json:"djzj"`
	Kqzj          int    `json:"kqzj"`
	Ljyk          int    `json:"ljyk"`
	Kyzj          int    `json:"kyzj"`
	MoneyType     string `json:"money_type" gorm:"column:money_type"`
	TotalsecMKval string `gorm:"column:totalsecMKval" json:"totalsecMKval"`
}

func (Summary) TableName() string {
	return "summary"
}

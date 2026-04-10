package model

type Role struct {
	BaseModel
	Name   string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Code   string `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Remark string `gorm:"size:256" json:"remark"`
	Status int8   `gorm:"default:1;comment:1=active,0=disabled" json:"status"`
}

func (Role) TableName() string {
	return "roles"
}

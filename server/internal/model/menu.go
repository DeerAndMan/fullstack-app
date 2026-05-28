package model

type Menu struct {
	ID       int64  `gorm:"column:id;type:bigint;primaryKey" json:"id"`
	Name     string `gorm:"column:name;type:varchar(100);not null" json:"name"`
	LinkURL  string `gorm:"column:link_url;type:varchar(500);not null" json:"link_url"`
	MenuCode string `gorm:"column:menu_code;type:varchar(100);not null" json:"menu_code"`
	ParentID int64  `gorm:"column:parent_id;type:bigint" json:"parent_id"`
	NodeType int8   `gorm:"column:node_type;type:tinyint" json:"node_type"`
	IconURL  string `gorm:"column:icon_url;type:varchar(255)" json:"icon_url"`
	Level    int    `gorm:"column:level;type:int;not null" json:"level"`
	Path     string `gorm:"column:path;type:varchar(2500)" json:"path"`
	IsDelete int8   `gorm:"column:is_delete;type:tinyint;not null" json:"is_delete"`
}

func (Menu) TableName() string {
	return "sys_menu"
}

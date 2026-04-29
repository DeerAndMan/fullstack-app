package model

import "time"

type SysMenuRole struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MenuID     int64     `gorm:"column:menu_id;not null" json:"menu_id"`
	RoleID     int64     `gorm:"column:role_id;not null" json:"role_id"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time" json:"update_time"`
	CreateUser int64     `gorm:"column:create_user;not null" json:"create_user"`
	UpdateUser int64     `gorm:"column:update_user" json:"update_user"`
	IsDelete   int8      `gorm:"column:is_delete;not null;type:tinyint(1)" json:"is_delete"`
}

func (SysMenuRole) TableName() string {
	return "sys_menu_role"
}

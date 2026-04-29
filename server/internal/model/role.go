package model

import "time"

type Role struct {
	RoleID     int64     `gorm:"column:role_id;primaryKey;autoIncrement" json:"role_id"`
	RoleName   string    `gorm:"column:role_name;type:varchar(64);not null" json:"role_name"`
	RoleKey    string    `gorm:"column:role_key;type:varchar(100);not null" json:"role_key"`
	Sort       int       `gorm:"column:sort;type:int;not null" json:"sort"`
	RoleStatus int8      `gorm:"column:role_status;type:tinyint;not null" json:"role_status"`
	CreateBy   string    `gorm:"column:create_by;type:varchar(64);not null" json:"create_by"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime;not null" json:"create_time"`
	UpdateBy   string    `gorm:"column:update_by;type:varchar(64);not null" json:"update_by"`
	UpdateTime time.Time `gorm:"column:update_time;type:datetime;not null" json:"update_time"`
	Remark     *string   `gorm:"column:remark;type:varchar(500)" json:"remark"`
	DelFlag    int8      `gorm:"column:del_flag;type:tinyint(1);not null" json:"del_flag"`
}

func (Role) TableName() string {
	return "sys_role"
}

package model

type SysUserRole struct {
	UserID int64 `gorm:"column:user_id;primaryKey" json:"user_id"`
	RoleID int64 `gorm:"column:role_id;primaryKey" json:"role_id"`
}

func (SysUserRole) TableName() string {
	return "sys_user_role"
}

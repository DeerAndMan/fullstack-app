package model

type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password string `gorm:"size:128;not null" json:"-"`
	Nickname string `gorm:"size:64" json:"nickname"`
	Avatar   string `gorm:"size:256" json:"avatar"`
	Email    string `gorm:"size:128" json:"email"`
	Phone    string `gorm:"size:20" json:"phone"`
	Status   int8   `gorm:"default:1;comment:1=active,0=disabled" json:"status"`
	Roles    []Role `gorm:"many2many:user_roles" json:"roles"`
}

func (User) TableName() string {
	return "users"
}

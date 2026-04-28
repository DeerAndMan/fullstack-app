package model

type User struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Password    string `gorm:"size:128;not null" json:"-"`
	Age         int8   `gorm:"type:tinyint" json:"age"`
	Email       string `gorm:"size:128" json:"email"`
	Description string `gorm:"size:10000" json:"description"`
	Status      int8   `gorm:"type:tinyint;default:1;not null;comment:1=active,0=disabled" json:"status"`
}

func (User) TableName() string {
	return "user"
}

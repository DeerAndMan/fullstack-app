package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Password    string         `gorm:"size:128;not null" json:"-"`
	Age         int            `json:"age"`
	Email       string         `gorm:"size:128" json:"email"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:datetime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:datetime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Description string         `gorm:"size:10000" json:"description"`
	Status      int8           `gorm:"type:tinyint;default:1;not null;comment:1=active,0=disabled" json:"status"`
}

func (User) TableName() string {
	return "user"
}

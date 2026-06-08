package model

import "time"

type XqThemeContent struct {
	ID           int64      `gorm:"primaryKey;not null;comment:随机生成的ID" json:"id"`
	UserID       int64      `gorm:"primaryKey;not null;comment:用户ID" json:"user_id"`
	ScreenName   string     `gorm:"type:varchar(100);comment:用户名" json:"screen_name"`
	CreatedAt    *time.Time `gorm:"type:timestamp;comment:主题创建时间" json:"created_at"`
	EditedAt     *time.Time `gorm:"type:timestamp;comment:主题编辑时间" json:"edited_at"`
	Text         string     `gorm:"type:mediumtext;comment:内容" json:"text"`
	MetaKeywords string     `gorm:"type:varchar(255);comment:地理信息" json:"meta_keywords"`
}

func (XqThemeContent) TableName() string {
	return "xq_theme_content"
}

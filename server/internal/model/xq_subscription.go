package model

type XqSubscription struct {
	ID          int64  `gorm:"primaryKey;not null;comment:订阅ID" json:"id"`
	UserID      int64  `gorm:"primaryKey;not null;comment:用户ID" json:"user_id"`
	Description string `gorm:"type:varchar(255);comment:描述" json:"description"`
	Enabled     *bool  `gorm:"type:tinyint(1);default:1;not null;comment:是否启用" json:"enabled"`
	IsStarred   bool   `gorm:"not null;default:false;comment:是否星标" json:"is_starred"`
	SortOrder   int32  `gorm:"not null;default:0;comment:排序值，越大越靠前" json:"sort_order"`
}

func (XqSubscription) TableName() string {
	return "xq_subscription"
}

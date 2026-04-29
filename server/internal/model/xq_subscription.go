package model

type XqSubscription struct {
	ID          int64  `gorm:"primaryKey;not null;comment:订阅ID" json:"id"`
	UserID      int64  `gorm:"primaryKey;not null;comment:用户ID" json:"user_id"`
	Description string `gorm:"type:varchar(255);comment:描述" json:"description"`
	Enabled     *bool  `gorm:"type:tinyint(1);default:1;not null;comment:是否启用" json:"enabled"`
}

func (XqSubscription) TableName() string {
	return "xq_subscription"
}

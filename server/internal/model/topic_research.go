package model

type TopicResearchSkill struct {
	BaseModel
	UserID       uint    `gorm:"not null;index;uniqueIndex:idx_user_skill_seed" json:"-"`
	SeedKey      *string `gorm:"size:64;uniqueIndex:idx_user_skill_seed" json:"-"`
	Name         string  `gorm:"size:80;not null" json:"name"`
	Instructions string  `gorm:"type:text;not null" json:"instructions"`
}

type TopicResearchRun struct {
	BaseModel
	UserID      uint   `gorm:"not null;index" json:"-"`
	Question    string `gorm:"size:255;not null" json:"question"`
	Model       string `gorm:"size:32;not null" json:"model"`
	Status      string `gorm:"size:20;not null" json:"status"`
	ReferenceID uint   `json:"referenceId,omitempty"`
	SkillText   string `gorm:"type:longtext;not null" json:"-"`
	ResultText  string `gorm:"type:longtext;not null" json:"-"`
	ErrorText   string `gorm:"size:255" json:"error,omitempty"`
}

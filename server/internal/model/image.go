package model

type Image struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	RelevanceID int    `gorm:"column:relevance_id" json:"relevance_id"`
	ImageName   string `gorm:"column:imageName;size:50" json:"image_name"`
	Image       []byte `gorm:"column:image;type:longblob" json:"image"`
	Size        int    `gorm:"column:size" json:"size"`
	UploadTime  string `gorm:"column:uploadTime;size:50" json:"upload_time"`
}

func (Image) TableName() string {
	return "image"
}

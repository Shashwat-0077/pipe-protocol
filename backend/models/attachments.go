package models

type Attachment struct {
	ID        uint `gorm:"primaryKey"`
	MessageID uint `gorm:"index;not null"`
	FileName  string
	MIMEType  string
	Size      int64
	Path      string // for S3

	Message Message `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
}

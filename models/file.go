package models

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID                   int        `gorm:"primaryKey;autoIncrement;type:int"`
	UploadTime           time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	FileType             string     `gorm:"size:50;not null"`
	FileSize             int64      `gorm:"type:bigint;not null"`
	AvailabilityDuration int        `gorm:"not null"`
	EncryptionKey        string     `gorm:"type:text;not null"`
	FileHash             string     `gorm:"size:64;not null"`
	OriginalFilename     string     `gorm:"size:255;not null"`
	StoredFilename       string     `gorm:"size:255;not null"`
	DownloadToken        string     `gorm:"size:64;uniqueIndex;not null"`
	ExpiryTime           time.Time  `gorm:"not null;index"`
	FileData             FileData   `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE"`
	Downloads            []Download `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE"`
}

func (f *File) BeforeCreate(tx *gorm.DB) (err error) {
	f.ExpiryTime = f.UploadTime.Add(time.Duration(f.AvailabilityDuration) * time.Minute)
	return
}

type FileData struct {
	ID     int    `gorm:"primaryKey;autoIncrement;type:int"`
	FileID int    `gorm:"not null;index;type:int"`
	Data   []byte `gorm:"type:longblob;not null"`
	IV     []byte `gorm:"type:varbinary(16);not null"`
	HMAC   []byte `gorm:"type:varbinary(32);not null"`
}

type Download struct {
	ID           int       `gorm:"primaryKey;autoIncrement;type:int"`
	FileID       int       `gorm:"not null;index;type:int"`
	DownloadTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	IPAddress    string    `gorm:"size:45;not null"`
}

type ScheduledEvent struct {
	ID            int    `gorm:"primaryKey;autoIncrement;type:int"`
	EventName     string `gorm:"uniqueIndex"`
	FileID        int
	ScheduledTime time.Time
	ExecutedTime  *time.Time
	ErrorMessage  *string
}

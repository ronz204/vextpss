package storages

import "time"

type SecretRecord struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	Name       string    `gorm:"uniqueIndex;not null"`
	Type       string    `gorm:"not null"`
	Algorithm  string    `gorm:"not null"`
	Ciphertext []byte    `gorm:"not null;type:blob"`
	Metadata   []byte    `gorm:"not null;type:blob"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

func (SecretRecord) TableName() string { return "secrets" }

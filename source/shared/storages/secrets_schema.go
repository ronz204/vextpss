package storages

import "time"

type SecretRecord struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	SpaceID    int64     `gorm:"not null;default:0;uniqueIndex:idx_space_name"`
	Name       string    `gorm:"not null;uniqueIndex:idx_space_name"`
	Type       string    `gorm:"not null"`
	Algorithm  string    `gorm:"not null"`
	Ciphertext []byte    `gorm:"not null;type:blob"`
	Metadata   []byte    `gorm:"not null;type:blob"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

func (SecretRecord) TableName() string { return "secrets" }

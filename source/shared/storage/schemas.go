package storage

import "time"

// ======================================
// SecretRecord is a 1:1 mapping of the `secrets` table.
// Encrypted is always opaque bytes — no plaintext ever persists here.
// ======================================
type SecretRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"uniqueIndex;not null"`
	Type      string    `gorm:"not null"`
	Salt      []byte    `gorm:"not null;type:blob"`
	Nonce     []byte    `gorm:"not null;type:blob"`
	Encrypted []byte    `gorm:"not null;type:blob"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (SecretRecord) TableName() string { return "secrets" }

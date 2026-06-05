package dal

import "time"

// SecretRecord is a 1:1 mapping of the `secrets` table.
// It contains no plaintext secret data — EncryptedPayload is always opaque bytes.
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

// TableName tells GORM to use "secrets" instead of the default "secret_records".
func (SecretRecord) TableName() string { return "secrets" }

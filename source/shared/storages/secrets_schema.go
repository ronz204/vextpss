package storages

import (
	"time"

	"vextpss/source/secrets/core"
)

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

func toRecord(s core.Secret, spaceID int64) SecretRecord {
	return SecretRecord{
		SpaceID:    spaceID,
		Name:       s.Name,
		Type:       s.Type,
		Algorithm:  s.Encrypted.Algorithm,
		Ciphertext: s.Encrypted.Ciphertext,
		Metadata:   s.Encrypted.Metadata,
	}
}

func toSecret(r SecretRecord, spaceName string) core.Secret {
	return core.Secret{
		ID:    r.ID,
		Space: spaceName,
		Name:  r.Name,
		Type:  r.Type,
		Encrypted: core.Encrypted{
			Algorithm:  r.Algorithm,
			Ciphertext: r.Ciphertext,
			Metadata:   r.Metadata,
		},
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

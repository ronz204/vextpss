package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"vextpss/source/secrets"
	"vextpss/source/shared/sentinel"

	"gorm.io/gorm"
)

type SecretRepository struct {
	db *gorm.DB
}

func NewSecretRepository(db *gorm.DB) *SecretRepository {
	return &SecretRepository{db: db}
}

// ================================
// Create persists a new secret with its pre-encrypted payload.
// ================================
func (r *SecretRepository) Create(ctx context.Context, secret *secrets.Secret, encrypted []byte) error {
	record := &SecretRecord{
		Name:      secret.Name,
		Type:      secret.Type,
		Salt:      secret.Salt,
		Nonce:     secret.Nonce,
		Encrypted: encrypted,
	}
	result := r.db.WithContext(ctx).Create(record)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return sentinel.ErrAlreadyExists
		}
		return fmt.Errorf("insert failed: %w", result.Error)
	}
	return nil
}

// ================================
// ListAll returns metadata for all secrets ordered by name.
// Blob columns (salt, nonce, encrypted) are excluded from the projection.
// ================================
func (r *SecretRepository) ListAll(ctx context.Context) ([]secrets.Secret, error) {
	var records []SecretRecord
	result := r.db.WithContext(ctx).
		Select("id, name, type, created_at, updated_at").
		Order("name asc").
		Find(&records)
	if result.Error != nil {
		return nil, fmt.Errorf("list query failed: %w", result.Error)
	}

	out := make([]secrets.Secret, len(records))
	for i, rec := range records {
		out[i] = secrets.Secret{
			ID:        secrets.SecretID(rec.ID),
			Name:      rec.Name,
			Type:      rec.Type,
			CreatedAt: rec.CreatedAt,
			UpdatedAt: rec.UpdatedAt,
		}
	}
	return out, nil
}

// ================================
// GetAll returns all secrets together with their encrypted payloads.
// ================================
func (r *SecretRepository) GetAll(ctx context.Context) ([]secrets.Credential, error) {
	var records []SecretRecord
	result := r.db.WithContext(ctx).Order("name asc").Find(&records)
	if result.Error != nil {
		return nil, fmt.Errorf("get all query failed: %w", result.Error)
	}

	out := make([]secrets.Credential, len(records))
	for i, rec := range records {
		out[i] = secrets.Credential{
			Secret: secrets.Secret{
				ID:        secrets.SecretID(rec.ID),
				Name:      rec.Name,
				Type:      rec.Type,
				Salt:      rec.Salt,
				Nonce:     rec.Nonce,
				CreatedAt: rec.CreatedAt,
				UpdatedAt: rec.UpdatedAt,
			},
			Encrypted: rec.Encrypted,
		}
	}
	return out, nil
}

// ================================
// GetByName retrieves a secret and its encrypted payload by name.
// Returns sentinel.ErrSecretNotFound when no row matches.
// ================================
func (r *SecretRepository) GetByName(ctx context.Context, name string) (*secrets.Secret, []byte, error) {
	var record SecretRecord
	result := r.db.WithContext(ctx).Where("name = ?", name).First(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil, sentinel.ErrSecretNotFound
		}
		return nil, nil, fmt.Errorf("get by name failed: %w", result.Error)
	}

	secret := &secrets.Secret{
		ID:        secrets.SecretID(record.ID),
		Name:      record.Name,
		Type:      record.Type,
		Salt:      record.Salt,
		Nonce:     record.Nonce,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
	return secret, record.Encrypted, nil
}

// ================================
// Update replaces the crypto material and encrypted payload for an existing secret.
// Returns sentinel.ErrSecretNotFound when no row matches.
// ================================
func (r *SecretRepository) Update(ctx context.Context, secret *secrets.Secret, encrypted []byte) error {
	result := r.db.WithContext(ctx).
		Model(&SecretRecord{}).
		Where("name = ?", secret.Name).
		Updates(map[string]any{
			"salt":      secret.Salt,
			"nonce":     secret.Nonce,
			"encrypted": encrypted,
		})
	if result.Error != nil {
		return fmt.Errorf("update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return sentinel.ErrSecretNotFound
	}
	return nil
}

// ================================
// Delete removes a secret by name.
// Returns sentinel.ErrSecretNotFound when no row matches.
// ================================
func (r *SecretRepository) Delete(ctx context.Context, name string) error {
	result := r.db.WithContext(ctx).Where("name = ?", name).Delete(&SecretRecord{})
	if result.Error != nil {
		return fmt.Errorf("delete failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return sentinel.ErrSecretNotFound
	}
	return nil
}

package repos

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"vextpss/source/core"
	"vextpss/source/dal"

	"gorm.io/gorm"
)

// SQLiteRepository implements SecretRepository using GORM + SQLite.
type SQLiteRepository struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Save(ctx context.Context, secret *core.Secret, encrypted []byte) error {
	record := &dal.SecretRecord{
		Name:      secret.Name,
		Type:      secret.Type,
		Salt:      secret.Salt,
		Nonce:     secret.Nonce,
		Encrypted: encrypted,
	}
	result := r.db.WithContext(ctx).Create(record)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return core.ErrAlreadyExists
		}
		return fmt.Errorf("insert failed: %w", result.Error)
	}
	return nil
}

func (r *SQLiteRepository) GetByName(ctx context.Context, name string) (*core.Secret, []byte, error) {
	var record dal.SecretRecord
	result := r.db.WithContext(ctx).Where("name = ?", name).First(&record)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil, core.ErrSecretNotFound
	}
	if result.Error != nil {
		return nil, nil, fmt.Errorf("query failed: %w", result.Error)
	}
	secret := &core.Secret{
		ID:        core.SecretID(record.ID),
		Name:      record.Name,
		Type:      record.Type,
		Salt:      record.Salt,
		Nonce:     record.Nonce,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
	return secret, record.Encrypted, nil
}

func (r *SQLiteRepository) ListAll(ctx context.Context) ([]core.Secret, error) {
	var records []dal.SecretRecord
	result := r.db.WithContext(ctx).
		Select("id, name, type, created_at, updated_at").
		Order("name asc").
		Find(&records)
	if result.Error != nil {
		return nil, fmt.Errorf("list query failed: %w", result.Error)
	}
	secrets := make([]core.Secret, len(records))
	for i, rec := range records {
		secrets[i] = core.Secret{
			ID:        core.SecretID(rec.ID),
			Name:      rec.Name,
			Type:      rec.Type,
			CreatedAt: rec.CreatedAt,
			UpdatedAt: rec.UpdatedAt,
		}
	}
	return secrets, nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, name string) error {
	result := r.db.WithContext(ctx).Where("name = ?", name).Delete(&dal.SecretRecord{})
	if result.Error != nil {
		return fmt.Errorf("delete failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return core.ErrSecretNotFound
	}
	return nil
}

func isUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

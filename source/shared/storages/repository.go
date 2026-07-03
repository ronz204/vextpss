package storages

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"vextpss/source/secrets"
)

var _ secrets.Repository = (*GORMRepository)(nil)

type GORMRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) Create(ctx context.Context, secret secrets.Secret) error {
	rec := toRecord(secret)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		if isDuplicate(err) {
			return secrets.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *GORMRepository) GetByName(ctx context.Context, name string) (secrets.Secret, error) {
	var rec SecretRecord
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return secrets.Secret{}, secrets.ErrNotFound
	}
	if err != nil {
		return secrets.Secret{}, err
	}
	return toSecret(rec), nil
}

func (r *GORMRepository) Update(ctx context.Context, secret secrets.Secret) error {
	result := r.db.WithContext(ctx).
		Model(&SecretRecord{}).
		Where("name = ?", secret.Name).
		Updates(map[string]any{
			"type":       secret.Type,
			"algorithm":  secret.Encrypted.Algorithm,
			"ciphertext": secret.Encrypted.Ciphertext,
			"metadata":   secret.Encrypted.Metadata,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return secrets.ErrNotFound
	}
	return nil
}

func (r *GORMRepository) Delete(ctx context.Context, name string) error {
	result := r.db.WithContext(ctx).Where("name = ?", name).Delete(&SecretRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return secrets.ErrNotFound
	}
	return nil
}

func (r *GORMRepository) List(ctx context.Context) ([]secrets.Secret, error) {
	var records []SecretRecord
	err := r.db.WithContext(ctx).
		Select("id, name, type, created_at, updated_at").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	items := make([]secrets.Secret, len(records))
	for i, rec := range records {
		items[i] = toSecret(rec)
	}
	return items, nil
}

func isDuplicate(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

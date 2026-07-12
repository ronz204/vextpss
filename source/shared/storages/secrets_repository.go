package storages

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"vextpss/source/secrets/core"
)

var _ core.SecretRepository = (*GORMRepository)(nil)

type GORMRepository struct {
	db *gorm.DB
}

func NewSecrets(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) resolveSpaceID(ctx context.Context, space string) (int64, error) {
	var rec SpaceRecord
	err := r.db.WithContext(ctx).Where("name = ?", space).First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, core.ErrSpaceNotFound
	}
	if err != nil {
		return 0, err
	}
	return rec.ID, nil
}

func (r *GORMRepository) Create(ctx context.Context, secret core.Secret) error {
	spaceID, err := r.resolveSpaceID(ctx, secret.Space)
	if err != nil {
		return err
	}
	rec := toRecord(secret, spaceID)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		if isDuplicate(err) {
			return core.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *GORMRepository) GetByName(ctx context.Context, space, name string) (core.Secret, error) {
	spaceID, err := r.resolveSpaceID(ctx, space)
	if err != nil {
		return core.Secret{}, err
	}
	var rec SecretRecord
	err = r.db.WithContext(ctx).
		Where("space_id = ? AND name = ?", spaceID, name).
		First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.Secret{}, core.ErrNotFound
	}
	if err != nil {
		return core.Secret{}, err
	}
	return toSecret(rec, space), nil
}

func (r *GORMRepository) Update(ctx context.Context, secret core.Secret) error {
	spaceID, err := r.resolveSpaceID(ctx, secret.Space)
	if err != nil {
		return err
	}
	result := r.db.WithContext(ctx).
		Model(&SecretRecord{}).
		Where("space_id = ? AND name = ?", spaceID, secret.Name).
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
		return core.ErrNotFound
	}
	return nil
}

func (r *GORMRepository) Rename(ctx context.Context, space, oldName, newName string) error {
	spaceID, err := r.resolveSpaceID(ctx, space)
	if err != nil {
		return err
	}
	result := r.db.WithContext(ctx).
		Model(&SecretRecord{}).
		Where("space_id = ? AND name = ?", spaceID, oldName).
		Updates(map[string]any{"name": newName})
	if result.Error != nil {
		if isDuplicate(result.Error) {
			return core.ErrAlreadyExists
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r *GORMRepository) Delete(ctx context.Context, space, name string) error {
	spaceID, err := r.resolveSpaceID(ctx, space)
	if err != nil {
		return err
	}
	result := r.db.WithContext(ctx).
		Where("space_id = ? AND name = ?", spaceID, name).
		Delete(&SecretRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r *GORMRepository) List(ctx context.Context, space string) ([]core.Secret, error) {
	var records []SecretRecord
	q := r.db.WithContext(ctx).
		Select("id, space_id, name, type, created_at, updated_at")

	if space != "" {
		spaceID, err := r.resolveSpaceID(ctx, space)
		if err != nil {
			return nil, err
		}
		if err := q.Where("space_id = ?", spaceID).Find(&records).Error; err != nil {
			return nil, err
		}
		items := make([]core.Secret, len(records))
		for i, rec := range records {
			items[i] = toSecret(rec, space)
		}
		return items, nil
	}

	if err := q.Find(&records).Error; err != nil {
		return nil, err
	}
	var spaceRecords []SpaceRecord
	if err := r.db.WithContext(ctx).Find(&spaceRecords).Error; err != nil {
		return nil, err
	}
	spaceNames := make(map[int64]string, len(spaceRecords))
	for _, s := range spaceRecords {
		spaceNames[s.ID] = s.Name
	}
	items := make([]core.Secret, len(records))
	for i, rec := range records {
		items[i] = toSecret(rec, spaceNames[rec.SpaceID])
	}
	return items, nil
}

func isDuplicate(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

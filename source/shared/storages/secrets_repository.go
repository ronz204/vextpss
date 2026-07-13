package storages

import (
	"context"
	"errors"

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
	rec, err := gorm.G[SpaceRecord](r.db).Where("name = ?", space).Take(ctx)
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
	if err := gorm.G[SecretRecord](r.db).Create(ctx, &rec); err != nil {
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
	rec, err := gorm.G[SecretRecord](r.db).
		Where("space_id = ? AND name = ?", spaceID, name).
		Take(ctx)
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
	n, err := gorm.G[SecretRecord](r.db).
		Where("space_id = ? AND name = ?", spaceID, oldName).
		Update(ctx, "name", newName)
	if err != nil {
		if isDuplicate(err) {
			return core.ErrAlreadyExists
		}
		return err
	}
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r *GORMRepository) Delete(ctx context.Context, space, name string) error {
	spaceID, err := r.resolveSpaceID(ctx, space)
	if err != nil {
		return err
	}
	n, err := gorm.G[SecretRecord](r.db).
		Where("space_id = ? AND name = ?", spaceID, name).
		Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r *GORMRepository) List(ctx context.Context, space string) ([]core.Secret, error) {
	if space != "" {
		spaceID, err := r.resolveSpaceID(ctx, space)
		if err != nil {
			return nil, err
		}
		records, err := gorm.G[SecretRecord](r.db).
			Select("id, space_id, name, type, created_at, updated_at").
			Where("space_id = ?", spaceID).Find(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]core.Secret, len(records))
		for i, rec := range records {
			items[i] = toSecret(rec, space)
		}
		return items, nil
	}

	records, err := gorm.G[SecretRecord](r.db).
		Select("id, space_id, name, type, created_at, updated_at").Find(ctx)
	if err != nil {
		return nil, err
	}
	spaceRecords, err := gorm.G[SpaceRecord](r.db).Find(ctx)
	if err != nil {
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
	type coder interface{ Code() int }
	var e coder
	return errors.As(err, &e) && e.Code() == 2067 // SQLITE_CONSTRAINT_UNIQUE
}

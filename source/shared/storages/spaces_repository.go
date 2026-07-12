package storages

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"vextpss/source/secrets/core"
)

var _ core.SpaceRepository = (*GORMSpaceRepository)(nil)

type GORMSpaceRepository struct {
	db *gorm.DB
}

func NewSpaces(db *gorm.DB) *GORMSpaceRepository {
	return &GORMSpaceRepository{db: db}
}

func (r *GORMSpaceRepository) Create(ctx context.Context, name string) error {
	rec := SpaceRecord{Name: name}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		if isDuplicate(err) {
			return core.ErrSpaceAlreadyExists
		}
		return err
	}
	return nil
}

func (r *GORMSpaceRepository) GetByName(ctx context.Context, name string) (core.Space, error) {
	var rec SpaceRecord
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.Space{}, core.ErrSpaceNotFound
	}
	if err != nil {
		return core.Space{}, err
	}
	return toSpace(rec), nil
}

func (r *GORMSpaceRepository) Rename(ctx context.Context, oldName, newName string) error {
	result := r.db.WithContext(ctx).
		Model(&SpaceRecord{}).
		Where("name = ?", oldName).
		Updates(map[string]any{"name": newName})
	if result.Error != nil {
		if isDuplicate(result.Error) {
			return core.ErrSpaceAlreadyExists
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return core.ErrSpaceNotFound
	}
	return nil
}

func (r *GORMSpaceRepository) Delete(ctx context.Context, name string) error {
	result := r.db.WithContext(ctx).Where("name = ?", name).Delete(&SpaceRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return core.ErrSpaceNotFound
	}
	return nil
}

func (r *GORMSpaceRepository) List(ctx context.Context) ([]core.Space, error) {
	var records []SpaceRecord
	if err := r.db.WithContext(ctx).Find(&records).Error; err != nil {
		return nil, err
	}
	spaces := make([]core.Space, len(records))
	for i, rec := range records {
		spaces[i] = toSpace(rec)
	}
	return spaces, nil
}

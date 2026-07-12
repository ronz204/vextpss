package storages

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"vextpss/source/secrets/core"
)

var _ core.StateRepository = (*GORMStateRepository)(nil)

const activeSpaceKey = "active:space"

type GORMStateRepository struct {
	db *gorm.DB
}

func NewState(db *gorm.DB) *GORMStateRepository {
	return &GORMStateRepository{db: db}
}

func (r *GORMStateRepository) GetActiveSpace(ctx context.Context) (string, error) {
	var rec MetaRecord
	err := r.db.WithContext(ctx).Where("key = ?", activeSpaceKey).First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return toActiveSpace(rec), nil
}

func (r *GORMStateRepository) SetActiveSpace(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).Save(&MetaRecord{
		Key:   activeSpaceKey,
		Value: name,
	}).Error
}

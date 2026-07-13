package storages

import (
	"context"

	"gorm.io/gorm"

	"vextpss/source/secrets/core"
)

var _ core.StateRepository = (*GORMStateRepository)(nil)

type GORMStateRepository struct {
	db *gorm.DB
}

func NewState(db *gorm.DB) *GORMStateRepository {
	return &GORMStateRepository{db: db}
}

func (r *GORMStateRepository) Load(ctx context.Context) (core.State, error) {
	var records []MetaRecord
	if err := r.db.WithContext(ctx).Find(&records).Error; err != nil {
		return core.State{}, err
	}
	return toState(records), nil
}

func (r *GORMStateRepository) Save(ctx context.Context, state core.State) error {
	for _, rec := range fromState(state) {
		if err := r.db.WithContext(ctx).Save(&rec).Error; err != nil {
			return err
		}
	}
	return nil
}

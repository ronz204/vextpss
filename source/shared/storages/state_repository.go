package storages

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	records, err := gorm.G[MetaRecord](r.db).Find(ctx)
	if err != nil {
		return core.State{}, err
	}
	return toState(records), nil
}

func (r *GORMStateRepository) Save(ctx context.Context, state core.State) error {
	records := fromState(state)
	return gorm.G[MetaRecord](r.db, clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).CreateInBatches(ctx, &records, len(records))
}

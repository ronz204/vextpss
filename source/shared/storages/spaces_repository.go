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
	if err := gorm.G[SpaceRecord](r.db).Create(ctx, &rec); err != nil {
		if isDuplicate(err) {
			return core.ErrSpaceAlreadyExists
		}
		return err
	}
	return nil
}

func (r *GORMSpaceRepository) GetByName(ctx context.Context, name string) (core.Space, error) {
	rec, err := gorm.G[SpaceRecord](r.db).Where("name = ?", name).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.Space{}, core.ErrSpaceNotFound
	}
	if err != nil {
		return core.Space{}, err
	}
	return toSpace(rec), nil
}

func (r *GORMSpaceRepository) Rename(ctx context.Context, oldName, newName string) error {
	n, err := gorm.G[SpaceRecord](r.db).Where("name = ?", oldName).Update(ctx, "name", newName)
	if err != nil {
		if isDuplicate(err) {
			return core.ErrSpaceAlreadyExists
		}
		return err
	}
	if n == 0 {
		return core.ErrSpaceNotFound
	}
	return nil
}

func (r *GORMSpaceRepository) Delete(ctx context.Context, name string) error {
	n, err := gorm.G[SpaceRecord](r.db).Where("name = ?", name).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrSpaceNotFound
	}
	return nil
}

func (r *GORMSpaceRepository) List(ctx context.Context) ([]core.Space, error) {
	records, err := gorm.G[SpaceRecord](r.db).Find(ctx)
	if err != nil {
		return nil, err
	}
	spaces := make([]core.Space, len(records))
	for i, rec := range records {
		spaces[i] = toSpace(rec)
	}
	return spaces, nil
}

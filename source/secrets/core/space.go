package core

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSpaceNotFound      = errors.New("space not found")
	ErrSpaceAlreadyExists = errors.New("space already exists")
)

type Space struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type SpaceRepository interface {
	Create(ctx context.Context, name string) error
	GetByName(ctx context.Context, name string) (Space, error)
	Rename(ctx context.Context, oldName, newName string) error
	Delete(ctx context.Context, name string) error
	List(ctx context.Context) ([]Space, error)
}

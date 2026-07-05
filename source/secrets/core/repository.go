package core

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("secret not found")
	ErrAlreadyExists = errors.New("secret already exists")
)

type Repository interface {
	Create(ctx context.Context, secret Secret) error
	GetByName(ctx context.Context, name string) (Secret, error)
	Update(ctx context.Context, secret Secret) error
	Rename(ctx context.Context, oldName, newName string) error
	Delete(ctx context.Context, name string) error
	List(ctx context.Context) ([]Secret, error)
}

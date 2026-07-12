package core

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("secret not found")
	ErrAlreadyExists = errors.New("secret already exists")
)

type Secret struct {
	ID        int64
	Space     string
	Name      string
	Type      string
	Encrypted Encrypted
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	TypeAccount = "account"
	TypeFinance = "finance"
)

type Payload interface {
	Type() string
	Display()
	Validate() error
}

type SecretRepository interface {
	Create(ctx context.Context, secret Secret) error
	GetByName(ctx context.Context, space, name string) (Secret, error)
	Update(ctx context.Context, secret Secret) error
	Rename(ctx context.Context, space, oldName, newName string) error
	Delete(ctx context.Context, space, name string) error
	List(ctx context.Context, space string) ([]Secret, error)
}

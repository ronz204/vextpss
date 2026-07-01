package secrets

import "context"

type Repository interface {
	Delete(ctx context.Context, name string) error
	Create(ctx context.Context, secret Secret) error
	Update(ctx context.Context, secret Secret) error
	ListAll(ctx context.Context) ([]Secret, error)
}

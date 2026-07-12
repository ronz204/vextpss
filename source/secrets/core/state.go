package core

import "context"

type StateRepository interface {
	GetActiveSpace(ctx context.Context) (string, error)
	SetActiveSpace(ctx context.Context, name string) error
}

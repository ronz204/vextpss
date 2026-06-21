package funcs

import (
	"context"
	"fmt"

	"vextpss/source/secrets"
)

type DeleteSecretDto struct {
	Name string
}

// ================================
// validate checks that the name field is present.
// ================================
func (d DeleteSecretDto) validate() error {
	if d.Name == "" {
		return fmt.Errorf("%w: name is required", secrets.ErrInvalidInput)
	}
	return nil
}

type DeleteSecretFunc struct {
	repo secrets.Repository
}

func NewDeleteSecretFunc(repo secrets.Repository) *DeleteSecretFunc {
	return &DeleteSecretFunc{repo: repo}
}

// ================================
// Run validates the DTO and permanently removes the secret.
// Returns secrets.ErrSecretNotFound if no secret matches the name.
// ================================
func (f *DeleteSecretFunc) Run(ctx context.Context, dto DeleteSecretDto) error {
	if err := dto.validate(); err != nil {
		return err
	}
	return f.repo.Delete(ctx, dto.Name)
}

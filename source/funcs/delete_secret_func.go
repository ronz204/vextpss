package funcs

import (
	"context"
	"fmt"

	"vextpss/source/shared/sentinel"
)

// DeleteSecretDto carries the inputs for the delete use case.
type DeleteSecretDto struct {
	Name string
}

// ================================
// validate checks that the name field is present.
// ================================
func (d DeleteSecretDto) validate() error {
	if d.Name == "" {
		return fmt.Errorf("%w: name is required", sentinel.ErrInvalidInput)
	}
	return nil
}

// DeleteSecretFunc orchestrates removing a persisted secret by name.
type DeleteSecretFunc struct {
	repo Repository
}

// ================================
// NewDeleteSecretFunc wires the use case with its persistence dependency.
// ================================
func NewDeleteSecretFunc(repo Repository) *DeleteSecretFunc {
	return &DeleteSecretFunc{repo: repo}
}

// ================================
// Run validates the DTO and permanently removes the secret.
// Returns sentinel.ErrSecretNotFound if no secret matches the name.
// ================================
func (f *DeleteSecretFunc) Run(ctx context.Context, dto DeleteSecretDto) error {
	if err := dto.validate(); err != nil {
		return err
	}
	return f.repo.Delete(ctx, dto.Name)
}

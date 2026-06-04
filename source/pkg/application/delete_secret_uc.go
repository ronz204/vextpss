package application

import (
	"context"
	"vextpss/source/pkg/domain"
)

// DeleteSecretUC is the use case for permanently removing a stored secret.
type DeleteSecretUC struct {
	repo domain.SecretRepository
}

func NewDeleteSecretUC(repo domain.SecretRepository) *DeleteSecretUC {
	return &DeleteSecretUC{repo: repo}
}

// Execute removes the secret identified by name. Returns ErrSecretNotFound if absent.
func (uc *DeleteSecretUC) Execute(ctx context.Context, name string) error {
	return uc.repo.Delete(ctx, name)
}

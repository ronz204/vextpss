package apps

import (
	"context"

	"vextpss/source/dal"
)

// DeleteSecretUC is the use case for permanently removing a stored secret.
type DeleteSecretUC struct {
	repo dal.SecretRepository
}

func NewDeleteSecretUC(repo dal.SecretRepository) *DeleteSecretUC {
	return &DeleteSecretUC{repo: repo}
}

// Execute removes the secret identified by name. Returns core.ErrSecretNotFound if absent.
func (uc *DeleteSecretUC) Execute(ctx context.Context, name string) error {
	return uc.repo.Delete(ctx, name)
}

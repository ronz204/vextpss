package app

import (
	"context"

	"vextpss/source/core"
)

// ListSecretsUC is the use case for listing all stored secrets (metadata only — no decryption).
type ListSecretsUC struct {
	repo core.SecretRepository
}

func NewListSecretsUC(repo core.SecretRepository) *ListSecretsUC {
	return &ListSecretsUC{repo: repo}
}

// Execute returns all secrets ordered by name. No master password is required.
func (uc *ListSecretsUC) Execute(ctx context.Context) ([]core.Secret, error) {
	return uc.repo.ListAll(ctx)
}

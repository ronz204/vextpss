package application

import (
	"context"
	"vextpss/source/pkg/domain"
)

// ListSecretsUC is the use case for listing all stored secrets (metadata only — no decryption).
type ListSecretsUC struct {
	repo domain.SecretRepository
}

func NewListSecretsUC(repo domain.SecretRepository) *ListSecretsUC {
	return &ListSecretsUC{repo: repo}
}

// Execute returns all secrets ordered by name. No master password is required.
func (uc *ListSecretsUC) Execute(ctx context.Context) ([]domain.Secret, error) {
	return uc.repo.ListAll(ctx)
}

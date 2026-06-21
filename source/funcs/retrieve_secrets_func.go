package funcs

import (
	"context"
	"fmt"

	"vextpss/source/secrets"
)

type RetrieveSecretsFunc struct {
	repo secrets.Repository
}

func NewRetrieveSecretsFunc(repo secrets.Repository) *RetrieveSecretsFunc {
	return &RetrieveSecretsFunc{repo: repo}
}

// ================================
// Run returns all secrets ordered by name. No master password required.
// ================================
func (f *RetrieveSecretsFunc) Run(ctx context.Context) ([]secrets.Secret, error) {
	all, err := f.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	return all, nil
}

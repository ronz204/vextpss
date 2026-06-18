package funcs

import (
	"context"
	"fmt"

	"vextpss/source/core"
	"vextpss/source/secrets"
)

// RetrieveSecretsFunc orchestrates listing secret metadata without touching encrypted payloads.
type RetrieveSecretsFunc struct {
	repo core.SecretRepository
}

// ================================
// NewRetrieveSecretsFunc wires the use case with its persistence dependency.
// ================================
func NewRetrieveSecretsFunc(repo core.SecretRepository) *RetrieveSecretsFunc {
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

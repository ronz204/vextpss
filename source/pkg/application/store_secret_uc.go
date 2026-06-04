package application

import (
	"context"
	"encoding/json"
	"fmt"
	"vextpss/source/pkg/domain"
	"vextpss/source/pkg/shared"
)

// StoreSecretUC is the use case for persisting a new secret.
type StoreSecretUC struct {
	repo      domain.SecretRepository
	encryptor domain.Encryptor
}

func NewStoreSecretUC(repo domain.SecretRepository, enc domain.Encryptor) *StoreSecretUC {
	return &StoreSecretUC{repo: repo, encryptor: enc}
}

// StoreSecretRequest carries all inputs from the interface layer (CLI, HTTP, etc.).
type StoreSecretRequest struct {
	Name            string
	Type            string // "account" — future: "card", "note"
	MasterPassword  []byte
	AccountUsername string
	AccountPassword []byte
}

func (r *StoreSecretRequest) validate() error {
	if r.Name == "" {
		return domain.NewDomainError("name is required")
	}
	if r.Type != "account" {
		return domain.NewDomainError("unsupported secret type")
	}
	if len(r.MasterPassword) == 0 {
		return domain.NewDomainError("master password is required")
	}
	return nil
}

func (r *StoreSecretRequest) buildPayload() (domain.SecretPayload, error) {
	switch r.Type {
	case "account":
		return &domain.AccountSecret{
			Username: r.AccountUsername,
			Password: r.AccountPassword,
		}, nil
	default:
		return nil, domain.NewDomainError("unknown secret type")
	}
}

// Execute validates inputs, encrypts the payload, and delegates persistence to the repo.
func (uc *StoreSecretUC) Execute(ctx context.Context, req StoreSecretRequest) error {
	if err := req.validate(); err != nil {
		return err
	}

	payload, err := req.buildPayload()
	if err != nil {
		return err
	}

	if err := payload.Validate(); err != nil {
		return err
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}
	defer shared.Zero(plaintext)

	salt, nonce, ciphertext, err := uc.encryptor.Encrypt(ctx, plaintext, req.MasterPassword)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	secret := &domain.Secret{
		Name:      req.Name,
		Type:      req.Type,
		Salt:      salt,
		Nonce:     nonce,
		CreatedAt: shared.Now(),
		UpdatedAt: shared.Now(),
	}

	if err := uc.repo.Save(ctx, secret, ciphertext); err != nil {
		if domain.IsDomainError(err) {
			return err
		}
		return fmt.Errorf("failed to store secret: %w", err)
	}

	return nil
}

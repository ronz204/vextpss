package apps

import (
	"context"
	"encoding/json"
	"fmt"

	"vextpss/source/core"
	"vextpss/source/dal"
	"vextpss/source/pkg/shared"
	"vextpss/source/pkg/tokens"
)

// StoreSecretUC is the use case for persisting a new secret.
type StoreSecretUC struct {
	repo      dal.SecretRepository
	encryptor tokens.Encryptor
}

func NewStoreSecretUC(repo dal.SecretRepository, enc tokens.Encryptor) *StoreSecretUC {
	return &StoreSecretUC{repo: repo, encryptor: enc}
}

// StoreSecretRequest carries all inputs from the interface layer (CLI, HTTP, etc.).
// The caller is responsible for building and providing the typed Payload.
type StoreSecretRequest struct {
	Name           string
	MasterPassword []byte
	Payload        core.SecretPayload
}

func (r *StoreSecretRequest) validate() error {
	if r.Name == "" {
		return core.NewDomainError("name is required")
	}
	if len(r.MasterPassword) == 0 {
		return core.NewDomainError("master password is required")
	}
	if r.Payload == nil {
		return core.NewDomainError("payload is required")
	}
	return nil
}

// Execute validates inputs, encrypts the payload, and delegates persistence to the repo.
func (uc *StoreSecretUC) Execute(ctx context.Context, req StoreSecretRequest) error {
	if err := req.validate(); err != nil {
		return err
	}

	if err := req.Payload.Validate(); err != nil {
		return err
	}

	plaintext, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}
	defer shared.Zero(plaintext)

	salt, nonce, ciphertext, err := uc.encryptor.Encrypt(ctx, plaintext, req.MasterPassword)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	secret := &core.Secret{
		Name:      req.Name,
		Type:      req.Payload.GetType(),
		Salt:      salt,
		Nonce:     nonce,
		CreatedAt: shared.Now(),
		UpdatedAt: shared.Now(),
	}

	if err := uc.repo.Save(ctx, secret, ciphertext); err != nil {
		if core.IsDomainError(err) {
			return err
		}
		return fmt.Errorf("failed to store secret: %w", err)
	}

	return nil
}

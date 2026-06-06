package apps

import (
	"context"
	"encoding/json"
	"fmt"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/dal"
	"vextpss/source/pkg/shared"
	"vextpss/source/pkg/tokens"
)

// UpdateSecretUC is the use case for replacing the payload of an existing secret.
type UpdateSecretUC struct {
	repo      dal.SecretRepository
	encryptor tokens.Encryptor
}

func NewUpdateSecretUC(repo dal.SecretRepository, enc tokens.Encryptor) *UpdateSecretUC {
	return &UpdateSecretUC{repo: repo, encryptor: enc}
}

// UpdateSecretRequest carries all inputs needed to update a stored credential.
// NewUsername may be empty — the UC will keep the existing username in that case.
type UpdateSecretRequest struct {
	Name           string
	MasterPassword []byte
	NewUsername    string
	NewPassword    []byte
}

func (r *UpdateSecretRequest) validate() error {
	if r.Name == "" {
		return core.NewDomainError("name is required")
	}
	if len(r.MasterPassword) == 0 {
		return core.NewDomainError("master password is required")
	}
	if len(r.NewPassword) == 0 {
		return core.NewDomainError("new password is required")
	}
	return nil
}

// Execute decrypts the current record, resolves the username, re-encrypts, and persists.
func (uc *UpdateSecretUC) Execute(ctx context.Context, req UpdateSecretRequest) error {
	if err := req.validate(); err != nil {
		return err
	}

	// Fetch existing record to verify existence and retrieve crypto material.
	secret, ciphertext, err := uc.repo.GetByName(ctx, req.Name)
	if err != nil {
		return err
	}

	// Decrypt current payload to resolve the username if no new one was provided.
	plaintext, err := uc.encryptor.Decrypt(ctx, req.MasterPassword, secret.Salt, secret.Nonce, ciphertext)
	if err != nil {
		return core.ErrDecryptionFailed
	}
	defer shared.Zero(plaintext)

	resolvedUsername := req.NewUsername
	if resolvedUsername == "" {
		currentPayload, err := deserialisePayload(secret.Type, plaintext)
		if err != nil {
			return err
		}
		if acc, ok := currentPayload.(*secrets.AccountSecret); ok {
			resolvedUsername = acc.Username
		}
	}

	// Build and validate the new payload.
	newPayload := &secrets.AccountSecret{
		Username: resolvedUsername,
		Password: req.NewPassword,
	}
	if err := newPayload.Validate(); err != nil {
		return err
	}

	newPlaintext, err := json.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}
	defer shared.Zero(newPlaintext)

	// Encrypt with fresh salt and nonce.
	newSalt, newNonce, newCiphertext, err := uc.encryptor.Encrypt(ctx, newPlaintext, req.MasterPassword)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	updated := &core.Secret{
		Name:      req.Name,
		Type:      secret.Type,
		Salt:      newSalt,
		Nonce:     newNonce,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: shared.Now(),
	}

	if err := uc.repo.Update(ctx, updated, newCiphertext); err != nil {
		if core.IsDomainError(err) {
			return err
		}
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

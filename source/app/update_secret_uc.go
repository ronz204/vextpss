package app

import (
	"context"
	"fmt"

	"vextpss/source/core"
	"vextpss/source/shared"
)

// UpdateSecretUC is the use case for replacing the payload of an existing secret.
type UpdateSecretUC struct {
	repo      core.SecretRepository
	encryptor core.Encryptor
}

func NewUpdateSecretUC(repo core.SecretRepository, enc core.Encryptor) *UpdateSecretUC {
	return &UpdateSecretUC{repo: repo, encryptor: enc}
}

// UpdateSecretRequest carries all inputs needed to update a stored credential.
// NewPayload must be of the same type as the stored secret.
type UpdateSecretRequest struct {
	Name           string
	MasterPassword []byte
	NewPayload     core.SecretPayload
}

func (r *UpdateSecretRequest) validate() error {
	if r.Name == "" {
		return core.NewDomainError("name is required")
	}
	if len(r.MasterPassword) == 0 {
		return core.NewDomainError("master password is required")
	}
	if r.NewPayload == nil {
		return core.NewDomainError("new payload is required")
	}
	return nil
}

// Execute verifies the master password, validates the new payload, re-encrypts, and persists.
func (uc *UpdateSecretUC) Execute(ctx context.Context, req UpdateSecretRequest) error {
	if err := req.validate(); err != nil {
		return err
	}

	secret, ciphertext, err := uc.repo.GetByName(ctx, req.Name)
	if err != nil {
		return err
	}

	// Decrypt current record solely to verify the master password is correct.
	plaintext, err := uc.encryptor.Decrypt(ctx, req.MasterPassword, secret.Salt, secret.Nonce, ciphertext)
	if err != nil {
		return core.ErrDecryptionFailed
	}
	defer shared.Zero(plaintext)

	// Guard against accidental type change.
	if req.NewPayload.GetType() != secret.Type {
		return core.NewDomainError(
			fmt.Sprintf("cannot change secret type from %q to %q", secret.Type, req.NewPayload.GetType()),
		)
	}

	if err := req.NewPayload.Validate(); err != nil {
		return err
	}

	newPlaintext, err := marshalPayload(req.NewPayload)
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

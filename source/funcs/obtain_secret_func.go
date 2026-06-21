package funcs

import (
	"context"
	"fmt"

	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
)

type ObtainSecretDto struct {
	Name           string
	MasterPassword []byte
}

// ================================
// validate checks that all required fields are present.
// ================================
func (d ObtainSecretDto) validate() error {
	switch {
	case d.Name == "":
		return fmt.Errorf("%w: name is required", secrets.ErrInvalidInput)
	case len(d.MasterPassword) == 0:
		return fmt.Errorf("%w: master password is required", secrets.ErrInvalidInput)
	}
	return nil
}

// ObtainSecretResult carries the decrypted payload and its type.
// Payload must be zeroed by the caller after use.
type ObtainSecretResult struct {
	Name    string
	Type    string
	Payload []byte
}

type ObtainSecretFunc struct {
	repo secrets.Repository
	enc  secrets.Encryptor
}

func NewObtainSecretFunc(repo secrets.Repository, enc secrets.Encryptor) *ObtainSecretFunc {
	return &ObtainSecretFunc{repo: repo, enc: enc}
}

// ================================
// Run validates, fetches the secret, and decrypts the payload.
// MasterPassword is zeroed on return. Payload in the result must be zeroed by the caller.
// ================================
func (f *ObtainSecretFunc) Run(ctx context.Context, dto ObtainSecretDto) (ObtainSecretResult, error) {
	defer memory.Cleaner(dto.MasterPassword)

	if err := dto.validate(); err != nil {
		return ObtainSecretResult{}, err
	}

	secret, encrypted, err := f.repo.GetByName(ctx, dto.Name)
	if err != nil {
		return ObtainSecretResult{}, err
	}

	plaintext, err := f.enc.Decrypt(ctx, dto.MasterPassword, secret.Salt, secret.Nonce, encrypted)
	if err != nil {
		return ObtainSecretResult{}, secrets.ErrDecryptionFailed
	}

	return ObtainSecretResult{
		Name:    secret.Name,
		Type:    secret.Type,
		Payload: plaintext,
	}, nil
}

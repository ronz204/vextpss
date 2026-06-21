package funcs

import (
	"context"
	"fmt"

	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
)

type UpdateSecretDto struct {
	Name           string
	Type           string
	Plaintext      []byte
	MasterPassword []byte
}

// ================================
// validate checks that all required fields are present and that Type is a known value.
// ================================
func (d UpdateSecretDto) validate() error {
	switch {
	case d.Name == "":
		return fmt.Errorf("%w: name is required", secrets.ErrInvalidInput)
	case !secrets.IsKnownType(d.Type):
		return fmt.Errorf("%w: type %q is not valid", secrets.ErrInvalidInput, d.Type)
	case len(d.Plaintext) == 0:
		return fmt.Errorf("%w: plaintext payload is required", secrets.ErrInvalidInput)
	case len(d.MasterPassword) == 0:
		return fmt.Errorf("%w: master password is required", secrets.ErrInvalidInput)
	}
	return nil
}

type UpdateSecretFunc struct {
	repo secrets.Repository
	enc  secrets.Encryptor
}

func NewUpdateSecretFunc(repo secrets.Repository, enc secrets.Encryptor) *UpdateSecretFunc {
	return &UpdateSecretFunc{repo: repo, enc: enc}
}

// ================================
// Run validates, encrypts the replacement payload, and persists the updated secret.
// Zeroes Plaintext and MasterPassword on return regardless of outcome.
// ================================
func (f *UpdateSecretFunc) Run(ctx context.Context, dto UpdateSecretDto) error {
	defer memory.Cleaner(dto.MasterPassword)
	defer memory.Cleaner(dto.Plaintext)

	if err := dto.validate(); err != nil {
		return err
	}

	salt, nonce, ciphertext, err := f.enc.Encrypt(ctx, dto.Plaintext, dto.MasterPassword)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	secret := &secrets.Secret{
		Name:  dto.Name,
		Type:  dto.Type,
		Salt:  salt,
		Nonce: nonce,
	}

	return f.repo.Update(ctx, secret, ciphertext)
}

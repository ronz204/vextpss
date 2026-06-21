package funcs

import (
	"context"
	"fmt"

	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
)

type DeleteSecretDto struct {
	Name           string
	MasterPassword []byte
}

func (d DeleteSecretDto) validate() error {
	if d.Name == "" {
		return fmt.Errorf("%w: name is required", secrets.ErrInvalidInput)
	}
	if len(d.MasterPassword) == 0 {
		return fmt.Errorf("%w: master password is required", secrets.ErrInvalidInput)
	}
	return nil
}

type DeleteSecretFunc struct {
	repo secrets.Repository
	enc  secrets.Encryptor
}

func NewDeleteSecretFunc(repo secrets.Repository, enc secrets.Encryptor) *DeleteSecretFunc {
	return &DeleteSecretFunc{repo: repo, enc: enc}
}

// Run verifies the master password by decrypting the secret before deleting it.
// Returns ErrDecryptionFailed if the password is wrong; ErrSecretNotFound if the name doesn't exist.
func (f *DeleteSecretFunc) Run(ctx context.Context, dto DeleteSecretDto) error {
	defer memory.Cleaner(dto.MasterPassword)

	if err := dto.validate(); err != nil {
		return err
	}

	secret, encrypted, err := f.repo.GetByName(ctx, dto.Name)
	if err != nil {
		return err
	}

	plaintext, err := f.enc.Decrypt(ctx, dto.MasterPassword, secret.Salt, secret.Nonce, encrypted)
	if err != nil {
		return secrets.ErrDecryptionFailed
	}
	defer memory.Cleaner(plaintext)

	return f.repo.Delete(ctx, dto.Name)
}

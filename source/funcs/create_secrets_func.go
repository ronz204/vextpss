package funcs

import (
	"context"
	"fmt"

	"vextpss/source/secrets"
	"vextpss/source/shared/cryptors"
	"vextpss/source/shared/memory"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

// CreateSecretDto carries the inputs for the create use case.
// Plaintext is the JSON-encoded secret payload — the caller is responsible for marshaling.
// Both Plaintext and MasterPassword are zeroed on Run return.
type CreateSecretDto struct {
	Name           string
	Type           string
	Plaintext      []byte
	MasterPassword []byte
}

// ================================
// validate checks that all required fields are present and that Type is a known value.
// ================================
func (d CreateSecretDto) validate() error {
	switch {
	case d.Name == "":
		return fmt.Errorf("%w: name is required", sentinel.ErrInvalidInput)
	case !secrets.IsKnownType(d.Type):
		return fmt.Errorf("%w: type %q is not valid", sentinel.ErrInvalidInput, d.Type)
	case len(d.Plaintext) == 0:
		return fmt.Errorf("%w: plaintext payload is required", sentinel.ErrInvalidInput)
	case len(d.MasterPassword) == 0:
		return fmt.Errorf("%w: master password is required", sentinel.ErrInvalidInput)
	}
	return nil
}

// CreateSecretFunc orchestrates encrypting and persisting a new secret.
type CreateSecretFunc struct {
	repo *storage.SecretRepository
	enc  *cryptors.AESGCMEncryptor
}

// ================================
// NewCreateSecretFunc wires the use case with its infrastructure dependencies.
// ================================
func NewCreateSecretFunc(repo *storage.SecretRepository, enc *cryptors.AESGCMEncryptor) *CreateSecretFunc {
	return &CreateSecretFunc{repo: repo, enc: enc}
}

// ================================
// Run validates, encrypts the plaintext payload, and persists the secret.
// Zeroes Plaintext and MasterPassword on return regardless of outcome.
// ================================
func (f *CreateSecretFunc) Run(ctx context.Context, dto CreateSecretDto) error {
	defer memory.Cleaner(dto.MasterPassword)
	defer memory.Cleaner(dto.Plaintext)

	if err := dto.validate(); err != nil {
		return err
	}

	out, err := f.enc.Encrypt(ctx, cryptors.EncryptInDto{
		Plaintext: dto.Plaintext,
		Password:  dto.MasterPassword,
	})
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	secret := &secrets.Secret{
		Name:  dto.Name,
		Type:  dto.Type,
		Salt:  out.Salt,
		Nonce: out.Nonce,
	}

	return f.repo.Create(ctx, secret, out.Ciphertext)
}

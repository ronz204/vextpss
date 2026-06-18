package funcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"vextpss/source/core"
	"vextpss/source/secrets"
	"vextpss/source/shared/passgen"
	"vextpss/source/shared/sentinel"
)

// ImportResult reports how many records were processed.
type ImportResult struct {
	Imported int
	Skipped  int // already-existing records that were left untouched
}

// ImportSecretsDto carries the inputs for the import use case.
type ImportSecretsDto struct {
	FilePath       string
	MasterPassword []byte
}

// ================================
// validate checks that all required fields are present.
// ================================
func (d ImportSecretsDto) validate() error {
	switch {
	case d.FilePath == "":
		return fmt.Errorf("%w: file path is required", sentinel.ErrInvalidInput)
	case len(d.MasterPassword) == 0:
		return fmt.Errorf("%w: master password is required", sentinel.ErrInvalidInput)
	}
	return nil
}

// ImportSecretsFunc orchestrates reading an encrypted export file and inserting its records.
type ImportSecretsFunc struct {
	repo core.SecretRepository
	enc  core.Encryptor
}

// ================================
// NewImportSecretsFunc wires the use case with its infrastructure dependencies.
// ================================
func NewImportSecretsFunc(repo core.SecretRepository, enc core.Encryptor) *ImportSecretsFunc {
	return &ImportSecretsFunc{repo: repo, enc: enc}
}

// ================================
// Run reads the export file, decrypts the bundle, and inserts each record.
// Records that already exist are skipped and counted in ImportResult.Skipped.
// MasterPassword is zeroed on return regardless of outcome.
// ================================
func (f *ImportSecretsFunc) Run(ctx context.Context, dto ImportSecretsDto) (ImportResult, error) {
	defer passgen.Cleaner(dto.MasterPassword)

	if err := dto.validate(); err != nil {
		return ImportResult{}, err
	}

	raw, err := os.ReadFile(dto.FilePath)
	if err != nil {
		return ImportResult{}, fmt.Errorf("read file: %w", err)
	}

	var ef exportFile
	if err = json.Unmarshal(raw, &ef); err != nil {
		return ImportResult{}, fmt.Errorf("parse export file: %w", err)
	}

	plaintext, err := f.enc.Decrypt(ctx, core.DecryptInDto{
		Password:   dto.MasterPassword,
		Salt:       ef.Salt,
		Nonce:      ef.Nonce,
		Ciphertext: ef.Data,
	})
	defer passgen.Cleaner(plaintext)
	if err != nil {
		return ImportResult{}, sentinel.ErrDecryptionFailed
	}

	var bundle exportBundle
	if err = json.Unmarshal(plaintext, &bundle); err != nil {
		return ImportResult{}, fmt.Errorf("parse bundle: %w", err)
	}

	var result ImportResult
	for _, rec := range bundle.Records {
		secret := &secrets.Secret{
			Name:  rec.Name,
			Type:  rec.Type,
			Salt:  rec.Salt,
			Nonce: rec.Nonce,
		}
		if err = f.repo.Create(ctx, secret, rec.Encrypted); err != nil {
			if errors.Is(err, sentinel.ErrAlreadyExists) {
				result.Skipped++
				continue
			}
			return result, fmt.Errorf("insert %q: %w", rec.Name, err)
		}
		result.Imported++
	}

	return result, nil
}

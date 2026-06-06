package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"vextpss/source/core"
	"vextpss/source/dal"
	"vextpss/source/pkg/shared"
	"vextpss/source/pkg/tokens"
)

// ImportSecretsUC imports secrets from a file produced by ExportSecretsUC.
type ImportSecretsUC struct {
	repo      dal.SecretRepository
	encryptor tokens.Encryptor
}

func NewImportSecretsUC(repo dal.SecretRepository, enc tokens.Encryptor) *ImportSecretsUC {
	return &ImportSecretsUC{repo: repo, encryptor: enc}
}

// ImportSecretsRequest carries all inputs needed to restore from an export file.
type ImportSecretsRequest struct {
	MasterPassword []byte
	InputPath      string
}

// ImportSecretsResponse reports how many secrets were imported or skipped.
type ImportSecretsResponse struct {
	Imported int
	Skipped  int
}

func (r *ImportSecretsRequest) validate() error {
	if len(r.MasterPassword) == 0 {
		return core.NewDomainError("master password is required")
	}
	if r.InputPath == "" {
		return core.NewDomainError("input path is required")
	}
	return nil
}

// Execute reads the export file, decrypts the bundle, and re-encrypts each secret
// individually before saving. Secrets whose name already exists are skipped.
func (uc *ImportSecretsUC) Execute(ctx context.Context, req ImportSecretsRequest) (*ImportSecretsResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	fileData, err := os.ReadFile(req.InputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read import file: %w", err)
	}

	var ef exportFile
	if err := json.Unmarshal(fileData, &ef); err != nil {
		return nil, core.NewDomainError("invalid export file format")
	}

	bundleJSON, err := uc.encryptor.Decrypt(ctx, req.MasterPassword, ef.Salt, ef.Nonce, ef.Payload)
	if err != nil {
		return nil, core.ErrDecryptionFailed
	}
	defer shared.Zero(bundleJSON)

	var bundle exportBundle
	if err := json.Unmarshal(bundleJSON, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse export bundle: %w", err)
	}

	resp := &ImportSecretsResponse{}
	for _, es := range bundle.Secrets {
		plaintext := []byte(es.Payload)
		salt, nonce, ciphertext, err := uc.encryptor.Encrypt(ctx, plaintext, req.MasterPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt %q: %w", es.Name, err)
		}

		secret := &core.Secret{
			Name:      es.Name,
			Type:      es.Type,
			Salt:      salt,
			Nonce:     nonce,
			CreatedAt: shared.Now(),
			UpdatedAt: shared.Now(),
		}

		if err := uc.repo.Save(ctx, secret, ciphertext); err != nil {
			if core.IsAlreadyExists(err) {
				resp.Skipped++
				continue
			}
			return nil, fmt.Errorf("failed to save %q: %w", es.Name, err)
		}
		resp.Imported++
	}

	return resp, nil
}

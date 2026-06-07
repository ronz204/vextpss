package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"vextpss/source/core"
	"vextpss/source/shared"
)

// exportedSecret is a single decrypted secret inside an export bundle.
type exportedSecret struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// exportBundle is the plaintext structure that gets encrypted into the export file.
type exportBundle struct {
	Version    int              `json:"version"`
	ExportedAt time.Time        `json:"exported_at"`
	Secrets    []exportedSecret `json:"secrets"`
}

// exportFile is the on-disk JSON envelope (salt + nonce + ciphertext).
type exportFile struct {
	Version int    `json:"version"`
	Salt    []byte `json:"salt"`
	Nonce   []byte `json:"nonce"`
	Payload []byte `json:"payload"`
}

// ExportSecretsUC exports all secrets as a single encrypted file.
type ExportSecretsUC struct {
	repo      core.SecretRepository
	encryptor core.Encryptor
}

func NewExportSecretsUC(repo core.SecretRepository, enc core.Encryptor) *ExportSecretsUC {
	return &ExportSecretsUC{repo: repo, encryptor: enc}
}

// ExportSecretsRequest carries all inputs needed to produce an export file.
type ExportSecretsRequest struct {
	MasterPassword []byte
	OutputPath     string
}

func (r *ExportSecretsRequest) validate() error {
	if len(r.MasterPassword) == 0 {
		return core.NewDomainError("master password is required")
	}
	if r.OutputPath == "" {
		return core.NewDomainError("output path is required")
	}
	return nil
}

// Execute decrypts every stored secret, bundles them, re-encrypts the bundle,
// and writes the export file to OutputPath. Returns the number of exported secrets.
func (uc *ExportSecretsUC) Execute(ctx context.Context, req ExportSecretsRequest) (int, error) {
	if err := req.validate(); err != nil {
		return 0, err
	}

	records, err := uc.repo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to read secrets: %w", err)
	}

	bundle := exportBundle{Version: 1, ExportedAt: shared.Now()}
	for _, rec := range records {
		plaintext, err := uc.encryptor.Decrypt(ctx, req.MasterPassword, rec.Secret.Salt, rec.Secret.Nonce, rec.Encrypted)
		if err != nil {
			return 0, core.ErrDecryptionFailed
		}
		// Copy before zeroing: json.RawMessage shares the same backing array as plaintext,
		// so zeroing plaintext would corrupt the bundle before serialization.
		payload := make([]byte, len(plaintext))
		copy(payload, plaintext)
		shared.Zero(plaintext)
		bundle.Secrets = append(bundle.Secrets, exportedSecret{
			Name:    rec.Secret.Name,
			Type:    rec.Secret.Type,
			Payload: json.RawMessage(payload),
		})
	}

	bundleJSON, err := json.Marshal(bundle)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize export bundle: %w", err)
	}
	defer shared.Zero(bundleJSON)

	salt, nonce, ciphertext, err := uc.encryptor.Encrypt(ctx, bundleJSON, req.MasterPassword)
	if err != nil {
		return 0, fmt.Errorf("failed to encrypt export: %w", err)
	}

	ef := exportFile{Version: 1, Salt: salt, Nonce: nonce, Payload: ciphertext}
	fileJSON, err := json.Marshal(ef)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize export file: %w", err)
	}

	if err := os.WriteFile(req.OutputPath, fileJSON, 0600); err != nil {
		return 0, fmt.Errorf("failed to write export file: %w", err)
	}

	return len(bundle.Secrets), nil
}

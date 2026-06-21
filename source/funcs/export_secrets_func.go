package funcs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
	"vextpss/source/shared/sentinel"
)

// exportRecord is the serialized form of one credential inside the bundle.
// The payload is kept in its already-encrypted form — no re-decryption required.
type exportRecord struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Salt      []byte `json:"salt"`
	Nonce     []byte `json:"nonce"`
	Encrypted []byte `json:"encrypted"`
}

// exportBundle is the plaintext payload that gets encrypted into the export file.
type exportBundle struct {
	Version    int            `json:"version"`
	ExportedAt time.Time      `json:"exported_at"`
	Records    []exportRecord `json:"records"`
}

// exportFile is the on-disk representation: crypto material + encrypted bundle.
type exportFile struct {
	Version int    `json:"version"`
	Salt    []byte `json:"salt"`
	Nonce   []byte `json:"nonce"`
	Data    []byte `json:"data"`
}

// ExportSecretsDto carries the inputs for the export use case.
type ExportSecretsDto struct {
	FilePath       string
	MasterPassword []byte
}

// ================================
// validate checks that all required fields are present.
// ================================
func (d ExportSecretsDto) validate() error {
	switch {
	case d.FilePath == "":
		return fmt.Errorf("%w: output path is required", sentinel.ErrInvalidInput)
	case len(d.MasterPassword) == 0:
		return fmt.Errorf("%w: master password is required", sentinel.ErrInvalidInput)
	}
	return nil
}

// ExportSecretsFunc orchestrates reading all secrets and writing an encrypted export file.
type ExportSecretsFunc struct {
	repo secrets.Repository
	enc  secrets.Encryptor
}

// ================================
// NewExportSecretsFunc wires the use case with its infrastructure dependencies.
// ================================
func NewExportSecretsFunc(repo secrets.Repository, enc secrets.Encryptor) *ExportSecretsFunc {
	return &ExportSecretsFunc{repo: repo, enc: enc}
}

// ================================
// Run reads all credentials, bundles them, encrypts the bundle, and writes the file.
// MasterPassword is zeroed on return regardless of outcome.
// ================================
func (f *ExportSecretsFunc) Run(ctx context.Context, dto ExportSecretsDto) error {
	defer memory.Cleaner(dto.MasterPassword)

	if err := dto.validate(); err != nil {
		return err
	}

	credentials, err := f.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("read credentials: %w", err)
	}

	records := make([]exportRecord, len(credentials))
	for i, c := range credentials {
		records[i] = exportRecord{
			Name:      c.Secret.Name,
			Type:      c.Secret.Type,
			Salt:      c.Secret.Salt,
			Nonce:     c.Secret.Nonce,
			Encrypted: c.Encrypted,
		}
	}

	bundle, err := json.Marshal(exportBundle{
		Version:    1,
		ExportedAt: time.Now().UTC(),
		Records:    records,
	})
	if err != nil {
		return fmt.Errorf("marshal bundle: %w", err)
	}
	defer memory.Cleaner(bundle)

	salt, nonce, data, err := f.enc.Encrypt(ctx, bundle, dto.MasterPassword)
	if err != nil {
		return fmt.Errorf("encrypt bundle: %w", err)
	}

	ef, err := json.Marshal(exportFile{
		Version: 1,
		Salt:    salt,
		Nonce:   nonce,
		Data:    data,
	})
	if err != nil {
		return fmt.Errorf("marshal export file: %w", err)
	}

	if err = os.WriteFile(dto.FilePath, ef, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

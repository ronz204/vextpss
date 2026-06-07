package app_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"vextpss/source/core"
	"vextpss/source/app"
	"vextpss/source/testutil"
)

func makeGetAllFn(records []core.FullRecord) func(_ context.Context) ([]core.FullRecord, error) {
	return func(_ context.Context) ([]core.FullRecord, error) {
		return records, nil
	}
}

func TestExportSecretsUC_Execute_WritesFile(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "backup.vext")

	records := []core.FullRecord{
		{
			Secret:    core.Secret{Name: "github", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
			Encrypted: encryptedAccountPayload("alice", "hunter2"),
		},
	}
	repo := &testutil.MockRepository{GetAllFn: makeGetAllFn(records)}
	uc := app.NewExportSecretsUC(repo, &testutil.MockEncryptor{})

	count, err := uc.Execute(context.Background(), app.ExportSecretsRequest{
		MasterPassword: []byte("master"),
		OutputPath:     outPath,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Execute() count = %d, want 1", count)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("export file not created: %v", err)
	}
}

func TestExportSecretsUC_Execute_EmptyVault(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "empty.vext")
	repo := &testutil.MockRepository{GetAllFn: makeGetAllFn(nil)}
	uc := app.NewExportSecretsUC(repo, &testutil.MockEncryptor{})

	count, err := uc.Execute(context.Background(), app.ExportSecretsRequest{
		MasterPassword: []byte("master"),
		OutputPath:     outPath,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Execute() count = %d, want 0", count)
	}
}

func TestExportSecretsUC_Execute_FileIsValidJSON(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "backup.vext")
	repo := &testutil.MockRepository{GetAllFn: makeGetAllFn(nil)}
	uc := app.NewExportSecretsUC(repo, &testutil.MockEncryptor{})

	uc.Execute(context.Background(), app.ExportSecretsRequest{ //nolint
		MasterPassword: []byte("master"),
		OutputPath:     outPath,
	})

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		t.Errorf("export file is not valid JSON: %v", err)
	}
}

func TestExportSecretsUC_Execute_WrongMasterPassword(t *testing.T) {
	records := []core.FullRecord{
		{
			Secret:    core.Secret{Name: "svc", Type: "account", Salt: []byte("s"), Nonce: []byte("n")},
			Encrypted: []byte("ciphertext"),
		},
	}
	repo := &testutil.MockRepository{GetAllFn: makeGetAllFn(records)}
	enc := &testutil.MockEncryptor{
		DecryptFn: func(_ context.Context, _ []byte, _ []byte, _ []byte, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	uc := app.NewExportSecretsUC(repo, enc)

	_, err := uc.Execute(context.Background(), app.ExportSecretsRequest{
		MasterPassword: []byte("wrong"),
		OutputPath:     filepath.Join(t.TempDir(), "out.vext"),
	})
	if err != core.ErrDecryptionFailed {
		t.Errorf("Execute() error = %v, want ErrDecryptionFailed", err)
	}
}

func TestExportSecretsUC_Execute_EmptyMasterPassword(t *testing.T) {
	uc := app.NewExportSecretsUC(&testutil.MockRepository{}, &testutil.MockEncryptor{})
	_, err := uc.Execute(context.Background(), app.ExportSecretsRequest{
		MasterPassword: []byte{},
		OutputPath:     "out.vext",
	})
	if !core.IsDomainError(err) {
		t.Errorf("Execute() error = %v, want DomainError", err)
	}
}

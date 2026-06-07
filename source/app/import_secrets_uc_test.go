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

// buildExportFile creates a .vext file on disk using the real ExportSecretsUC
// so ImportSecretsUC tests operate on a genuine export.
func buildExportFile(t *testing.T, records []core.FullRecord, masterPw []byte) string {
	t.Helper()
	outPath := filepath.Join(t.TempDir(), "export.vext")
	repo := &testutil.MockRepository{GetAllFn: makeGetAllFn(records)}
	exportUC := app.NewExportSecretsUC(repo, &testutil.MockEncryptor{})
	if _, err := exportUC.Execute(context.Background(), app.ExportSecretsRequest{
		MasterPassword: masterPw,
		OutputPath:     outPath,
	}); err != nil {
		t.Fatalf("buildExportFile: %v", err)
	}
	return outPath
}

func TestImportSecretsUC_Execute_Success(t *testing.T) {
	records := []core.FullRecord{
		{
			Secret:    core.Secret{Name: "github", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
			Encrypted: encryptedAccountPayload("alice", "hunter2"),
		},
		{
			Secret:    core.Secret{Name: "netflix", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
			Encrypted: encryptedAccountPayload("bob", "n3tflix"),
		},
	}
	exportPath := buildExportFile(t, records, []byte("master"))

	repo := &testutil.MockRepository{}
	uc := app.NewImportSecretsUC(repo, &testutil.MockEncryptor{})

	resp, err := uc.Execute(context.Background(), app.ImportSecretsRequest{
		MasterPassword: []byte("master"),
		InputPath:      exportPath,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.Imported != 2 {
		t.Errorf("Imported = %d, want 2", resp.Imported)
	}
	if resp.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0", resp.Skipped)
	}
}

func TestImportSecretsUC_Execute_SkipsExisting(t *testing.T) {
	records := []core.FullRecord{
		{
			Secret:    core.Secret{Name: "github", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
			Encrypted: encryptedAccountPayload("alice", "pass"),
		},
	}
	exportPath := buildExportFile(t, records, []byte("master"))

	repo := &testutil.MockRepository{
		SaveFn: func(_ context.Context, _ *core.Secret, _ []byte) error {
			return core.ErrAlreadyExists
		},
	}
	uc := app.NewImportSecretsUC(repo, &testutil.MockEncryptor{})

	resp, err := uc.Execute(context.Background(), app.ImportSecretsRequest{
		MasterPassword: []byte("master"),
		InputPath:      exportPath,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.Imported != 0 {
		t.Errorf("Imported = %d, want 0", resp.Imported)
	}
	if resp.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", resp.Skipped)
	}
}

func TestImportSecretsUC_Execute_WrongMasterPassword(t *testing.T) {
	exportPath := buildExportFile(t, nil, []byte("master"))

	enc := &testutil.MockEncryptor{
		DecryptFn: func(_ context.Context, _ []byte, _ []byte, _ []byte, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	uc := app.NewImportSecretsUC(&testutil.MockRepository{}, enc)

	_, err := uc.Execute(context.Background(), app.ImportSecretsRequest{
		MasterPassword: []byte("wrong"),
		InputPath:      exportPath,
	})
	if err != core.ErrDecryptionFailed {
		t.Errorf("Execute() error = %v, want ErrDecryptionFailed", err)
	}
}

func TestImportSecretsUC_Execute_InvalidFile(t *testing.T) {
	badPath := filepath.Join(t.TempDir(), "bad.vext")
	os.WriteFile(badPath, []byte("not json"), 0600) //nolint

	uc := app.NewImportSecretsUC(&testutil.MockRepository{}, &testutil.MockEncryptor{})
	_, err := uc.Execute(context.Background(), app.ImportSecretsRequest{
		MasterPassword: []byte("master"),
		InputPath:      badPath,
	})
	if !core.IsDomainError(err) {
		t.Errorf("Execute() with invalid file error = %v, want DomainError", err)
	}
}

func TestImportSecretsUC_Execute_MissingFile(t *testing.T) {
	uc := app.NewImportSecretsUC(&testutil.MockRepository{}, &testutil.MockEncryptor{})
	_, err := uc.Execute(context.Background(), app.ImportSecretsRequest{
		MasterPassword: []byte("master"),
		InputPath:      "/nonexistent/path.vext",
	})
	if err == nil {
		t.Fatal("Execute() with missing file should return error")
	}
}

func TestImportSecretsUC_Execute_EmptyMasterPassword(t *testing.T) {
	uc := app.NewImportSecretsUC(&testutil.MockRepository{}, &testutil.MockEncryptor{})
	_, err := uc.Execute(context.Background(), app.ImportSecretsRequest{
		MasterPassword: []byte{},
		InputPath:      "some.vext",
	})
	if !core.IsDomainError(err) {
		t.Errorf("Execute() error = %v, want DomainError", err)
	}
}

func TestImportExportRoundtrip(t *testing.T) {
	records := []core.FullRecord{
		{
			Secret:    core.Secret{Name: "github", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
			Encrypted: encryptedAccountPayload("alice", "hunter2"),
		},
	}
	exportPath := buildExportFile(t, records, []byte("master"))

	var savedSecret *core.Secret
	var savedCiphertext []byte
	repo := &testutil.MockRepository{
		SaveFn: func(_ context.Context, s *core.Secret, enc []byte) error {
			savedSecret = s
			savedCiphertext = enc
			return nil
		},
	}
	uc := app.NewImportSecretsUC(repo, &testutil.MockEncryptor{})

	resp, err := uc.Execute(context.Background(), app.ImportSecretsRequest{
		MasterPassword: []byte("master"),
		InputPath:      exportPath,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.Imported != 1 {
		t.Fatalf("Imported = %d, want 1", resp.Imported)
	}
	if savedSecret.Name != "github" {
		t.Errorf("imported name = %q, want %q", savedSecret.Name, "github")
	}

	// The re-encrypted ciphertext (via MockEncryptor identity) should round-trip to the original payload.
	var ap struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(savedCiphertext, &ap); err != nil {
		t.Fatalf("unmarshal re-encrypted payload: %v", err)
	}
	if ap.Username != "alice" {
		t.Errorf("imported username = %q, want %q", ap.Username, "alice")
	}
}

package handlers_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/dal"
	"vextpss/source/pkg/apps"
	"vextpss/source/tests/mocks"
)

func newImportHandler(repo *mocks.MockRepository, enc *mocks.MockEncryptor, prompter helpers.Prompter) *handlers.ImportHandler {
	uc := apps.NewImportSecretsUC(repo, enc)
	return handlers.NewImportHandler(uc, prompter)
}

func makeExportFile(t *testing.T, records []dal.FullRecord, masterPw []byte) string {
	t.Helper()
	outPath := filepath.Join(t.TempDir(), "export.vext")
	exportRepo := &mocks.MockRepository{
		GetAllFn: func(_ context.Context) ([]dal.FullRecord, error) { return records, nil },
	}
	exportUC := apps.NewExportSecretsUC(exportRepo, &mocks.MockEncryptor{})
	if _, err := exportUC.Execute(context.Background(), apps.ExportSecretsRequest{
		MasterPassword: masterPw,
		OutputPath:     outPath,
	}); err != nil {
		t.Fatalf("makeExportFile: %v", err)
	}
	return outPath
}

func TestImportHandler_Handle_Success(t *testing.T) {
	records := []dal.FullRecord{
		{
			Secret:    core.Secret{Name: "github", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
			Encrypted: encryptedAccount("alice", "pass"),
		},
		{
			Secret:    core.Secret{Name: "netflix", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
			Encrypted: encryptedAccount("bob", "pass2"),
		},
	}
	importPath := makeExportFile(t, records, []byte("master"))

	prompter := &helpers.MockPrompter{PasswordBytes: []byte("master")}
	h := newImportHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), importPath); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if !strings.Contains(out, "[✓]") {
		t.Errorf("output %q should contain success marker [✓]", out)
	}
	if !strings.Contains(out, "2 secret") {
		t.Errorf("output %q should mention count", out)
	}
}

func TestImportHandler_Handle_SkipsExisting(t *testing.T) {
	records := []dal.FullRecord{
		{
			Secret:    core.Secret{Name: "github", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
			Encrypted: encryptedAccount("alice", "pass"),
		},
	}
	importPath := makeExportFile(t, records, []byte("master"))

	repo := &mocks.MockRepository{
		SaveFn: func(_ context.Context, _ *core.Secret, _ []byte) error {
			return core.ErrAlreadyExists
		},
	}
	prompter := &helpers.MockPrompter{PasswordBytes: []byte("master")}
	h := newImportHandler(repo, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		h.Handle(context.Background(), importPath) //nolint
	})

	if !strings.Contains(out, "skipped") {
		t.Errorf("output %q should mention skipped", out)
	}
}

func TestImportHandler_Handle_InvalidFile(t *testing.T) {
	badPath := filepath.Join(t.TempDir(), "bad.vext")
	os.WriteFile(badPath, []byte("not json"), 0600) //nolint

	prompter := &helpers.MockPrompter{PasswordBytes: []byte("master")}
	h := newImportHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		h.Handle(context.Background(), badPath) //nolint
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

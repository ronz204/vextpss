package handlers_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/ui"
	"vextpss/source/core"
	"vextpss/source/app"
	"vextpss/source/testutil"
)

func newImportHandler(repo *testutil.MockRepository, enc *testutil.MockEncryptor, prompter ui.Prompter) *handlers.ImportHandler {
	uc := app.NewImportSecretsUC(repo, enc)
	return handlers.NewImportHandler(uc, prompter)
}

func makeExportFile(t *testing.T, records []core.FullRecord, masterPw []byte) string {
	t.Helper()
	outPath := filepath.Join(t.TempDir(), "export.vext")
	exportRepo := &testutil.MockRepository{
		GetAllFn: func(_ context.Context) ([]core.FullRecord, error) { return records, nil },
	}
	exportUC := app.NewExportSecretsUC(exportRepo, &testutil.MockEncryptor{})
	if _, err := exportUC.Execute(context.Background(), app.ExportSecretsRequest{
		MasterPassword: masterPw,
		OutputPath:     outPath,
	}); err != nil {
		t.Fatalf("makeExportFile: %v", err)
	}
	return outPath
}

func TestImportHandler_Handle_Success(t *testing.T) {
	records := []core.FullRecord{
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

	prompter := &testutil.MockPrompter{PasswordBytes: []byte("master")}
	h := newImportHandler(&testutil.MockRepository{}, &testutil.MockEncryptor{}, prompter)

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
	records := []core.FullRecord{
		{
			Secret:    core.Secret{Name: "github", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
			Encrypted: encryptedAccount("alice", "pass"),
		},
	}
	importPath := makeExportFile(t, records, []byte("master"))

	repo := &testutil.MockRepository{
		SaveFn: func(_ context.Context, _ *core.Secret, _ []byte) error {
			return core.ErrAlreadyExists
		},
	}
	prompter := &testutil.MockPrompter{PasswordBytes: []byte("master")}
	h := newImportHandler(repo, &testutil.MockEncryptor{}, prompter)

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

	prompter := &testutil.MockPrompter{PasswordBytes: []byte("master")}
	h := newImportHandler(&testutil.MockRepository{}, &testutil.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		h.Handle(context.Background(), badPath) //nolint
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

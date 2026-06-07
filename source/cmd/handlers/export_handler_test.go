package handlers_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/ui"
	"vextpss/source/core"
	"vextpss/source/app"
	"vextpss/source/testutil"
)

func newExportHandler(repo *testutil.MockRepository, enc *testutil.MockEncryptor, prompter ui.Prompter) *handlers.ExportHandler {
	uc := app.NewExportSecretsUC(repo, enc)
	return handlers.NewExportHandler(uc, prompter)
}

func TestExportHandler_Handle_Success(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "backup.vext")
	repo := &testutil.MockRepository{
		GetAllFn: func(_ context.Context) ([]core.FullRecord, error) {
			return []core.FullRecord{
				{
					Secret:    core.Secret{Name: "github", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
					Encrypted: encryptedAccount("alice", "pass"),
				},
			}, nil
		},
	}
	prompter := &testutil.MockPrompter{PasswordBytes: []byte("masterpass")}
	h := newExportHandler(repo, &testutil.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), outPath); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if !strings.Contains(out, "[✓]") {
		t.Errorf("output %q should contain success marker [✓]", out)
	}
	if !strings.Contains(out, "1 secret") {
		t.Errorf("output %q should mention count", out)
	}
}

func TestExportHandler_Handle_WrongMasterPassword(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "backup.vext")
	repo := &testutil.MockRepository{
		GetAllFn: func(_ context.Context) ([]core.FullRecord, error) {
			return []core.FullRecord{
				{
					Secret:    core.Secret{Name: "svc", Type: "account", Salt: []byte("s"), Nonce: []byte("n")},
					Encrypted: []byte("ct"),
				},
			}, nil
		},
	}
	enc := &testutil.MockEncryptor{
		DecryptFn: func(_ context.Context, _ []byte, _ []byte, _ []byte, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	prompter := &testutil.MockPrompter{PasswordBytes: []byte("wrong")}
	h := newExportHandler(repo, enc, prompter)

	out := captureStdout(t, func() {
		h.Handle(context.Background(), outPath) //nolint
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

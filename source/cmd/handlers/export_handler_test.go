package handlers_test

import (
	"context"
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

func newExportHandler(repo *mocks.MockRepository, enc *mocks.MockEncryptor, prompter helpers.Prompter) *handlers.ExportHandler {
	uc := apps.NewExportSecretsUC(repo, enc)
	return handlers.NewExportHandler(uc, prompter)
}

func TestExportHandler_Handle_Success(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "backup.vext")
	repo := &mocks.MockRepository{
		GetAllFn: func(_ context.Context) ([]dal.FullRecord, error) {
			return []dal.FullRecord{
				{
					Secret:    core.Secret{Name: "github", Type: "account", Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b")},
					Encrypted: encryptedAccount("alice", "pass"),
				},
			}, nil
		},
	}
	prompter := &helpers.MockPrompter{PasswordBytes: []byte("masterpass")}
	h := newExportHandler(repo, &mocks.MockEncryptor{}, prompter)

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
	repo := &mocks.MockRepository{
		GetAllFn: func(_ context.Context) ([]dal.FullRecord, error) {
			return []dal.FullRecord{
				{
					Secret:    core.Secret{Name: "svc", Type: "account", Salt: []byte("s"), Nonce: []byte("n")},
					Encrypted: []byte("ct"),
				},
			}, nil
		},
	}
	enc := &mocks.MockEncryptor{
		DecryptFn: func(_ context.Context, _ []byte, _ []byte, _ []byte, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	prompter := &helpers.MockPrompter{PasswordBytes: []byte("wrong")}
	h := newExportHandler(repo, enc, prompter)

	out := captureStdout(t, func() {
		h.Handle(context.Background(), outPath) //nolint
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

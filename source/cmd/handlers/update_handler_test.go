package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/pkg/apps"
	"vextpss/source/tests/mocks"
)

func newUpdateHandler(repo *mocks.MockRepository, enc *mocks.MockEncryptor, prompter helpers.Prompter) *handlers.UpdateHandler {
	uc := apps.NewUpdateSecretUC(repo, enc)
	return handlers.NewUpdateHandler(uc, prompter)
}

func encryptedAccount(username, password string) []byte {
	b, _ := json.Marshal(&secrets.AccountSecret{Username: username, Password: []byte(password)})
	return b
}

func TestUpdateHandler_Handle_Success(t *testing.T) {
	repo := &mocks.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return &core.Secret{Name: "github", Type: "account",
				Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b"),
			}, encryptedAccount("alice", "old"), nil
		},
	}
	prompter := &sequentialPrompter{
		lineResponses: []string{"bob"},
		passwordResponses: []passwordResponse{
			{[]byte("new-pass"), nil},
			{[]byte("masterpass"), nil},
		},
	}
	h := newUpdateHandler(repo, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "github"); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if !strings.Contains(out, "[✓]") {
		t.Errorf("output %q should contain success marker [✓]", out)
	}
	if !strings.Contains(out, "github") {
		t.Errorf("output %q should contain credential name", out)
	}
}

func TestUpdateHandler_Handle_NotFound(t *testing.T) {
	repo := &mocks.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return nil, nil, core.ErrSecretNotFound
		},
	}
	prompter := &sequentialPrompter{
		lineResponses: []string{""},
		passwordResponses: []passwordResponse{
			{[]byte("new-pass"), nil},
			{[]byte("master"), nil},
		},
	}
	h := newUpdateHandler(repo, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		h.Handle(context.Background(), "missing") //nolint
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

func TestUpdateHandler_Handle_ReadLineError(t *testing.T) {
	prompter := &helpers.MockPrompter{Err: errors.New("stdin closed")}
	h := newUpdateHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	err := h.Handle(context.Background(), "svc")
	if err == nil {
		t.Fatal("Handle() should return error when ReadLine fails")
	}
}

func TestUpdateHandler_Handle_WrongMasterPassword(t *testing.T) {
	repo := &mocks.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return &core.Secret{Name: "svc", Type: "account",
				Salt: []byte("stub-salt-16byte"), Nonce: []byte("stub-nonce-12b"),
			}, encryptedAccount("alice", "old"), nil
		},
	}
	enc := &mocks.MockEncryptor{
		DecryptFn: func(_ context.Context, _ []byte, _ []byte, _ []byte, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	prompter := &sequentialPrompter{
		lineResponses: []string{""},
		passwordResponses: []passwordResponse{
			{[]byte("new-pass"), nil},
			{[]byte("wrong"), nil},
		},
	}
	h := newUpdateHandler(repo, enc, prompter)

	out := captureStdout(t, func() {
		h.Handle(context.Background(), "svc") //nolint
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X] on wrong master password", out)
	}
}

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/ui"
	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/app"
	"vextpss/source/testutil"
)

func newUpdateHandler(repo *testutil.MockRepository, enc *testutil.MockEncryptor, prompter ui.Prompter) *handlers.UpdateHandler {
	retrieveUC := app.NewRetrieveSecretUC(repo, enc)
	updateUC := app.NewUpdateSecretUC(repo, enc)
	return handlers.NewUpdateHandler(retrieveUC, updateUC, prompter)
}

func encryptedAccount(username, password string) []byte {
	b, _ := json.Marshal(&secrets.AccountSecret{Username: username, Password: []byte(password)})
	return b
}

func accountGetByNameFn(username, password string) func(_ context.Context, _ string) (*core.Secret, []byte, error) {
	return func(_ context.Context, _ string) (*core.Secret, []byte, error) {
		return &core.Secret{
			Name:  "github",
			Type:  secrets.TypeAccount,
			Salt:  []byte("stub-salt-16byte"),
			Nonce: []byte("stub-nonce-12b"),
		}, encryptedAccount(username, password), nil
	}
}

// TestUpdateHandler_Handle_Success verifies the golden path:
// master password → retrieve succeeds → collector gathers new values → update succeeds.
//
// Prompt order after refactoring:
//  1. ReadPassword("Master Password:")     → masterpass
//  2. ReadLine("Username:")                → "bob"    (AccountCollector)
//  3. ReadPassword("Password:")            → new-pass (AccountCollector)
func TestUpdateHandler_Handle_Success(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: accountGetByNameFn("alice", "old"),
	}
	prompter := &sequentialPrompter{
		passwordResponses: []passwordResponse{
			{[]byte("masterpass"), nil}, // master password
			{[]byte("new-pass"), nil},   // new password (from AccountCollector)
		},
		lineResponses: []string{"bob"}, // new username (from AccountCollector)
	}
	h := newUpdateHandler(repo, &testutil.MockEncryptor{}, prompter)

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
	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return nil, nil, core.ErrSecretNotFound
		},
	}
	prompter := &sequentialPrompter{
		passwordResponses: []passwordResponse{
			{[]byte("master"), nil}, // master password
		},
	}
	h := newUpdateHandler(repo, &testutil.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		h.Handle(context.Background(), "missing") //nolint
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

func TestUpdateHandler_Handle_ReadPasswordError(t *testing.T) {
	prompter := &testutil.MockPrompter{Err: errors.New("stdin closed")}
	h := newUpdateHandler(&testutil.MockRepository{}, &testutil.MockEncryptor{}, prompter)

	err := h.Handle(context.Background(), "svc")
	if err == nil {
		t.Fatal("Handle() should return error when ReadPassword fails")
	}
}

func TestUpdateHandler_Handle_WrongMasterPassword(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: accountGetByNameFn("alice", "old"),
	}
	enc := &testutil.MockEncryptor{
		DecryptFn: func(_ context.Context, _ []byte, _ []byte, _ []byte, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	prompter := &sequentialPrompter{
		passwordResponses: []passwordResponse{
			{[]byte("wrong"), nil}, // master password
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

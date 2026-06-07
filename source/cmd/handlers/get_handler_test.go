package handlers_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"vextpss/source/app"
	"vextpss/source/cmd/handlers"
	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/testutil"
)

func accountJSON(t *testing.T, username string, password []byte) []byte {
	t.Helper()
	b, err := json.Marshal(&secrets.AccountSecret{Username: username, Password: password})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return b
}

func TestGetHandler_Handle_Success(t *testing.T) {
	payload := accountJSON(t, "alice", []byte("s3cr3t"))
	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, name string) (*core.Secret, []byte, error) {
			return &core.Secret{Name: name, Type: "account", CreatedAt: time.Now()}, payload, nil
		},
	}
	prompter := &testutil.MockPrompter{PasswordBytes: []byte("masterpass")}
	uc := app.NewRetrieveSecretUC(repo, &testutil.MockEncryptor{})
	h := handlers.NewGetHandler(uc, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "github"); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if !strings.Contains(out, "alice") {
		t.Errorf("output %q should contain username 'alice'", out)
	}
	if !strings.Contains(out, "github") {
		t.Errorf("output %q should contain service 'github'", out)
	}
}

func TestGetHandler_Handle_NotFound(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return nil, nil, core.ErrSecretNotFound
		},
	}
	prompter := &testutil.MockPrompter{PasswordBytes: []byte("master")}
	uc := app.NewRetrieveSecretUC(repo, &testutil.MockEncryptor{})
	h := handlers.NewGetHandler(uc, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "missing"); err != nil {
			t.Errorf("Handle() should not return error, got %v", err)
		}
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
	if !strings.Contains(out, "missing") {
		t.Errorf("output %q should mention the secret name", out)
	}
}

func TestGetHandler_Handle_DecryptionFailed(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, name string) (*core.Secret, []byte, error) {
			return &core.Secret{Name: name, Type: "account"}, []byte("enc"), nil
		},
	}
	enc := &testutil.MockEncryptor{
		DecryptFn: func(_ context.Context, _, _, _, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	prompter := &testutil.MockPrompter{PasswordBytes: []byte("wrong")}
	uc := app.NewRetrieveSecretUC(repo, enc)
	h := handlers.NewGetHandler(uc, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "svc"); err != nil {
			t.Errorf("Handle() should not return error, got %v", err)
		}
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

func TestGetHandler_Handle_PromptError(t *testing.T) {
	prompter := &sequentialPrompter{
		passwordResponses: []passwordResponse{
			{nil, context.DeadlineExceeded},
		},
	}
	uc := app.NewRetrieveSecretUC(&testutil.MockRepository{}, &testutil.MockEncryptor{})
	h := handlers.NewGetHandler(uc, prompter)

	err := h.Handle(context.Background(), "svc")
	if err == nil {
		t.Fatal("Handle() should return error when ReadPassword fails")
	}
}

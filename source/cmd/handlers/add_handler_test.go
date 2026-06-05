package handlers_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/pkg/apps"
	"vextpss/source/tests/mocks"
)

func newAddHandler(repo *mocks.MockRepository, enc *mocks.MockEncryptor, prompter helpers.Prompter) *handlers.AddHandler {
	uc := apps.NewStoreSecretUC(repo, enc)
	return handlers.NewAddHandler(uc, prompter)
}

func TestAddHandler_Handle_Success(t *testing.T) {
	prompter := &helpers.MockPrompter{
		LineResponse:  "alice",
		PasswordBytes: []byte("s3cr3t"),
	}
	h := newAddHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "github"); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if !strings.Contains(out, "github") {
		t.Errorf("output %q should contain credential name", out)
	}
	if !strings.Contains(out, "[✓]") {
		t.Errorf("output %q should contain success marker [✓]", out)
	}
}

func TestAddHandler_Handle_AlreadyExists(t *testing.T) {
	prompter := &helpers.MockPrompter{
		LineResponse:  "alice",
		PasswordBytes: []byte("pass"),
	}
	repo := &mocks.MockRepository{
		SaveFn: func(_ context.Context, _ *core.Secret, _ []byte) error {
			return core.ErrAlreadyExists
		},
	}
	h := newAddHandler(repo, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "github"); err != nil {
			t.Errorf("Handle() should not return error, got %v", err)
		}
	})

	if !strings.Contains(out, "already exists") {
		t.Errorf("output %q should mention already exists", out)
	}
}

func TestAddHandler_Handle_DomainValidationError(t *testing.T) {
	// Empty username triggers domain validation error inside the UC.
	prompter := &helpers.MockPrompter{
		LineResponse:  "",
		PasswordBytes: []byte("pass"),
	}
	h := newAddHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "svc"); err != nil {
			t.Errorf("Handle() should not return error, got %v", err)
		}
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

func TestAddHandler_Handle_ReadLineError(t *testing.T) {
	prompter := &helpers.MockPrompter{
		Err: errors.New("stdin closed"),
	}
	h := newAddHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	err := h.Handle(context.Background(), "svc")
	if err == nil {
		t.Fatal("Handle() should return error when ReadLine fails")
	}
}

func TestAddHandler_Handle_ReadPasswordError(t *testing.T) {
	// ReadLine succeeds but the first ReadPassword call fails.
	prompter := &sequentialPrompter{
		lineResponses: []string{"alice"},
		passwordResponses: []passwordResponse{
			{nil, errors.New("no tty")},
		},
	}
	h := newAddHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	err := h.Handle(context.Background(), "svc")
	if err == nil {
		t.Fatal("Handle() should return error when ReadPassword fails")
	}
}

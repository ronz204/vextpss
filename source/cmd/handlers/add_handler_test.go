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
		if err := h.Handle(context.Background(), "github", "account", false, 20, true); err != nil {
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
		if err := h.Handle(context.Background(), "github", "account", false, 20, true); err != nil {
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
		if err := h.Handle(context.Background(), "svc", "account", false, 20, true); err != nil {
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

	err := h.Handle(context.Background(), "svc", "account", false, 20, true)
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

	err := h.Handle(context.Background(), "svc", "account", false, 20, true)
	if err == nil {
		t.Fatal("Handle() should return error when ReadPassword fails")
	}
}

func TestAddHandler_Handle_Generate(t *testing.T) {
	prompter := &helpers.MockPrompter{
		LineResponse:  "alice",
		PasswordBytes: []byte("masterpass"),
	}
	h := newAddHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "twitter", "account", true, 20, true); err != nil {
			t.Errorf("Handle() with --generate error = %v", err)
		}
	})

	if !strings.Contains(out, "Generated password") {
		t.Errorf("output %q should show generated password", out)
	}
	if !strings.Contains(out, "[✓]") {
		t.Errorf("output %q should contain success marker [✓]", out)
	}
}

func TestAddHandler_Handle_CreditSuccess(t *testing.T) {
	// Provide sequential responses for all credit card fields + master password.
	prompter := &sequentialPrompter{
		lineResponses: []string{
			"4532123456789012", // card number
			"6",               // expiration month
			"2028",            // expiration year
			"",                // bank username (skip)
			"",                // cellphone (skip)
			"",                // country code (skip)
		},
		passwordResponses: []passwordResponse{
			{[]byte("123"), nil},        // security code
			{[]byte("1234"), nil},       // PIN
			{[]byte("vk123"), nil},      // virtual key
			{[]byte("masterpass"), nil}, // master password
		},
	}
	h := newAddHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "visa-debit", "credit", false, 0, false); err != nil {
			t.Errorf("Handle() credit error = %v", err)
		}
	})

	if !strings.Contains(out, "[✓]") {
		t.Errorf("output %q should contain success marker [✓]", out)
	}
	if !strings.Contains(out, "visa-debit") {
		t.Errorf("output %q should contain credential name", out)
	}
}

func TestAddHandler_Handle_UnknownType(t *testing.T) {
	prompter := &helpers.MockPrompter{}
	h := newAddHandler(&mocks.MockRepository{}, &mocks.MockEncryptor{}, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "svc", "note", false, 0, false); err != nil {
			t.Errorf("Handle() should not return error for unknown type, got %v", err)
		}
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

package handlers_test

import (
	"context"
	"strings"
	"testing"

	"vextpss/source/app"
	"vextpss/source/cmd/handlers"
	"vextpss/source/core"
	"vextpss/source/testutil"
)

func TestRmHandler_Handle_ConfirmedSuccess(t *testing.T) {
	var deletedName string
	repo := &testutil.MockRepository{
		DeleteFn: func(_ context.Context, name string) error {
			deletedName = name
			return nil
		},
	}
	prompter := &testutil.MockPrompter{ConfirmResponse: true}
	uc := app.NewDeleteSecretUC(repo)
	h := handlers.NewRmHandler(uc, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "github"); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if deletedName != "github" {
		t.Errorf("deleted name = %q, want %q", deletedName, "github")
	}
	if !strings.Contains(out, "[✓]") {
		t.Errorf("output %q should contain success marker [✓]", out)
	}
}

func TestRmHandler_Handle_Rejected(t *testing.T) {
	var deleteCalled bool
	repo := &testutil.MockRepository{
		DeleteFn: func(_ context.Context, _ string) error {
			deleteCalled = true
			return nil
		},
	}
	prompter := &testutil.MockPrompter{ConfirmResponse: false}
	uc := app.NewDeleteSecretUC(repo)
	h := handlers.NewRmHandler(uc, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "github"); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if deleteCalled {
		t.Error("Delete should not be called when user rejects confirmation")
	}
	if !strings.Contains(out, "Aborted") {
		t.Errorf("output %q should say Aborted", out)
	}
}

func TestRmHandler_Handle_NotFound(t *testing.T) {
	repo := &testutil.MockRepository{
		DeleteFn: func(_ context.Context, _ string) error {
			return core.ErrSecretNotFound
		},
	}
	prompter := &testutil.MockPrompter{ConfirmResponse: true}
	uc := app.NewDeleteSecretUC(repo)
	h := handlers.NewRmHandler(uc, prompter)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background(), "missing"); err != nil {
			t.Errorf("Handle() should not return error, got %v", err)
		}
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

func TestRmHandler_Handle_ConfirmError(t *testing.T) {
	prompter := &sequentialPrompter{
		confirmResponse: false,
		confirmErr:      context.DeadlineExceeded,
	}
	uc := app.NewDeleteSecretUC(&testutil.MockRepository{})
	h := handlers.NewRmHandler(uc, prompter)

	err := h.Handle(context.Background(), "svc")
	if err == nil {
		t.Fatal("Handle() should return error when Confirm fails")
	}
}

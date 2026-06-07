package handlers_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"vextpss/source/cmd/handlers"
	"vextpss/source/core"
	"vextpss/source/app"
	"vextpss/source/testutil"
)

func TestListHandler_Handle_WithSecrets(t *testing.T) {
	repo := &testutil.MockRepository{
		ListAllFn: func(_ context.Context) ([]core.Secret, error) {
			return []core.Secret{
				{Name: "github", Type: "account", CreatedAt: time.Now()},
				{Name: "gitlab", Type: "account", CreatedAt: time.Now()},
			}, nil
		},
	}
	uc := app.NewListSecretsUC(repo)
	h := handlers.NewListHandler(uc)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background()); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if !strings.Contains(out, "github") {
		t.Errorf("output %q should contain 'github'", out)
	}
	if !strings.Contains(out, "gitlab") {
		t.Errorf("output %q should contain 'gitlab'", out)
	}
}

func TestListHandler_Handle_Empty(t *testing.T) {
	repo := &testutil.MockRepository{
		ListAllFn: func(_ context.Context) ([]core.Secret, error) {
			return []core.Secret{}, nil
		},
	}
	uc := app.NewListSecretsUC(repo)
	h := handlers.NewListHandler(uc)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background()); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if !strings.Contains(out, "No secrets") {
		t.Errorf("output %q should mention no secrets", out)
	}
}

func TestListHandler_Handle_RepoError(t *testing.T) {
	repo := &testutil.MockRepository{
		ListAllFn: func(_ context.Context) ([]core.Secret, error) {
			return nil, errors.New("db down")
		},
	}
	uc := app.NewListSecretsUC(repo)
	h := handlers.NewListHandler(uc)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background()); err != nil {
			t.Errorf("Handle() should not return error, got %v", err)
		}
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

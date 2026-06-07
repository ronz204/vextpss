package app_test

import (
	"context"
	"errors"
	"testing"

	"vextpss/source/core"
	"vextpss/source/app"
	"vextpss/source/testutil"
)

func TestDeleteSecretUC_Execute_Success(t *testing.T) {
	var deleted string
	repo := &testutil.MockRepository{
		DeleteFn: func(_ context.Context, name string) error {
			deleted = name
			return nil
		},
	}
	uc := app.NewDeleteSecretUC(repo)

	if err := uc.Execute(context.Background(), "github"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if deleted != "github" {
		t.Errorf("deleted name = %q, want %q", deleted, "github")
	}
}

func TestDeleteSecretUC_Execute_NotFound(t *testing.T) {
	repo := &testutil.MockRepository{
		DeleteFn: func(_ context.Context, _ string) error {
			return core.ErrSecretNotFound
		},
	}
	uc := app.NewDeleteSecretUC(repo)

	err := uc.Execute(context.Background(), "missing")
	if !core.IsNotFound(err) {
		t.Errorf("Execute() error = %v, want ErrSecretNotFound", err)
	}
}

func TestDeleteSecretUC_Execute_RepoError(t *testing.T) {
	repo := &testutil.MockRepository{
		DeleteFn: func(_ context.Context, _ string) error {
			return errors.New("disk full")
		},
	}
	uc := app.NewDeleteSecretUC(repo)

	err := uc.Execute(context.Background(), "svc")
	if err == nil {
		t.Fatal("Execute() with repo error should return an error")
	}
}

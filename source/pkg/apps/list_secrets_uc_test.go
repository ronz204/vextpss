package apps_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"vextpss/source/core"
	"vextpss/source/pkg/apps"
	"vextpss/source/tests/mocks"
)

func TestListSecretsUC_Execute_ReturnsAll(t *testing.T) {
	want := []core.Secret{
		{Name: "github", Type: "account", CreatedAt: time.Now()},
		{Name: "gitlab", Type: "account", CreatedAt: time.Now()},
	}
	repo := &mocks.MockRepository{
		ListAllFn: func(_ context.Context) ([]core.Secret, error) {
			return want, nil
		},
	}
	uc := apps.NewListSecretsUC(repo)

	got, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("len(secrets) = %d, want %d", len(got), len(want))
	}
	for i, s := range got {
		if s.Name != want[i].Name {
			t.Errorf("secrets[%d].Name = %q, want %q", i, s.Name, want[i].Name)
		}
	}
}

func TestListSecretsUC_Execute_EmptyList(t *testing.T) {
	repo := &mocks.MockRepository{
		ListAllFn: func(_ context.Context) ([]core.Secret, error) {
			return []core.Secret{}, nil
		},
	}
	uc := apps.NewListSecretsUC(repo)

	got, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d items", len(got))
	}
}

func TestListSecretsUC_Execute_RepoError(t *testing.T) {
	repo := &mocks.MockRepository{
		ListAllFn: func(_ context.Context) ([]core.Secret, error) {
			return nil, errors.New("db timeout")
		},
	}
	uc := apps.NewListSecretsUC(repo)

	_, err := uc.Execute(context.Background())
	if err == nil {
		t.Fatal("Execute() with repo error should return an error")
	}
}

package apps_test

import (
	"context"
	"encoding/json"
	"testing"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/pkg/apps"
	"vextpss/source/tests/mocks"
)

func encryptedAccountPayload(username, password string) []byte {
	payload := &secrets.AccountSecret{Username: username, Password: []byte(password)}
	b, _ := json.Marshal(payload)
	return b
}

func makeGetByNameFn(username, password string) func(_ context.Context, _ string) (*core.Secret, []byte, error) {
	return func(_ context.Context, _ string) (*core.Secret, []byte, error) {
		return &core.Secret{
			Name: "github",
			Type: "account",
			Salt: []byte("stub-salt-16byte"),
			Nonce: []byte("stub-nonce-12b"),
		}, encryptedAccountPayload(username, password), nil
	}
}

func TestUpdateSecretUC_Execute_Success(t *testing.T) {
	repo := &mocks.MockRepository{
		GetByNameFn: makeGetByNameFn("alice", "old-pass"),
	}
	uc := apps.NewUpdateSecretUC(repo, &mocks.MockEncryptor{})

	req := apps.UpdateSecretRequest{
		Name:           "github",
		MasterPassword: []byte("master"),
		NewUsername:    "bob",
		NewPassword:    []byte("new-pass"),
	}
	if err := uc.Execute(context.Background(), req); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestUpdateSecretUC_Execute_KeepsCurrentUsername(t *testing.T) {
	var savedPayload []byte
	repo := &mocks.MockRepository{
		GetByNameFn: makeGetByNameFn("alice", "old-pass"),
		UpdateFn: func(_ context.Context, _ *core.Secret, encrypted []byte) error {
			savedPayload = encrypted
			return nil
		},
	}
	uc := apps.NewUpdateSecretUC(repo, &mocks.MockEncryptor{})

	req := apps.UpdateSecretRequest{
		Name:           "github",
		MasterPassword: []byte("master"),
		NewUsername:    "", // empty — should keep "alice"
		NewPassword:    []byte("new-pass"),
	}
	if err := uc.Execute(context.Background(), req); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var ap secrets.AccountSecret
	if err := json.Unmarshal(savedPayload, &ap); err != nil {
		t.Fatalf("unmarshal saved payload: %v", err)
	}
	if ap.Username != "alice" {
		t.Errorf("kept username = %q, want %q", ap.Username, "alice")
	}
}

func TestUpdateSecretUC_Execute_NotFound(t *testing.T) {
	repo := &mocks.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return nil, nil, core.ErrSecretNotFound
		},
	}
	uc := apps.NewUpdateSecretUC(repo, &mocks.MockEncryptor{})

	req := apps.UpdateSecretRequest{
		Name:           "missing",
		MasterPassword: []byte("master"),
		NewPassword:    []byte("pass"),
	}
	err := uc.Execute(context.Background(), req)
	if !core.IsNotFound(err) {
		t.Errorf("Execute() error = %v, want ErrSecretNotFound", err)
	}
}

func TestUpdateSecretUC_Execute_WrongMasterPassword(t *testing.T) {
	repo := &mocks.MockRepository{
		GetByNameFn: makeGetByNameFn("alice", "old"),
	}
	enc := &mocks.MockEncryptor{
		DecryptFn: func(_ context.Context, _ []byte, _ []byte, _ []byte, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	uc := apps.NewUpdateSecretUC(repo, enc)

	req := apps.UpdateSecretRequest{
		Name:           "github",
		MasterPassword: []byte("wrong"),
		NewPassword:    []byte("pass"),
	}
	err := uc.Execute(context.Background(), req)
	if err != core.ErrDecryptionFailed {
		t.Errorf("Execute() error = %v, want ErrDecryptionFailed", err)
	}
}

func TestUpdateSecretUC_Execute_EmptyName(t *testing.T) {
	uc := apps.NewUpdateSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := apps.UpdateSecretRequest{MasterPassword: []byte("master"), NewPassword: []byte("pass")}
	err := uc.Execute(context.Background(), req)
	if !core.IsDomainError(err) {
		t.Errorf("Execute() with empty name should return DomainError, got %v", err)
	}
}

func TestUpdateSecretUC_Execute_EmptyNewPassword(t *testing.T) {
	uc := apps.NewUpdateSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := apps.UpdateSecretRequest{Name: "svc", MasterPassword: []byte("master")}
	err := uc.Execute(context.Background(), req)
	if !core.IsDomainError(err) {
		t.Errorf("Execute() with empty new password should return DomainError, got %v", err)
	}
}

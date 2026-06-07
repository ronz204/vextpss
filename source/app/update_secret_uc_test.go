package app_test

import (
	"context"
	"encoding/json"
	"testing"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/app"
	"vextpss/source/testutil"
)

// encryptedAccountPayload marshals an AccountSecret as the mock encryptor would store it.
func encryptedAccountPayload(username, password string) []byte {
	payload := &secrets.AccountSecret{Username: username, Password: []byte(password)}
	b, _ := json.Marshal(payload)
	return b
}

func makeGetByNameFn(secretType, username, password string) func(_ context.Context, _ string) (*core.Secret, []byte, error) {
	return func(_ context.Context, _ string) (*core.Secret, []byte, error) {
		return &core.Secret{
			Name:  "github",
			Type:  secretType,
			Salt:  []byte("stub-salt-16byte"),
			Nonce: []byte("stub-nonce-12b"),
		}, encryptedAccountPayload(username, password), nil
	}
}

func TestUpdateSecretUC_Execute_Success(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: makeGetByNameFn(secrets.TypeAccount, "alice", "old-pass"),
	}
	uc := app.NewUpdateSecretUC(repo, &testutil.MockEncryptor{})

	req := app.UpdateSecretRequest{
		Name:           "github",
		MasterPassword: []byte("master"),
		NewPayload:     &secrets.AccountSecret{Username: "bob", Password: []byte("new-pass")},
	}
	if err := uc.Execute(context.Background(), req); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestUpdateSecretUC_Execute_CreditSecret(t *testing.T) {
	creditPayload := &secrets.CreditSecret{
		Number:          "4532123456789012",
		SecurityCode:    []byte("123"),
		ExpirationMonth: 6,
		ExpirationYear:  2028,
		Pin:             []byte("1234"),
	}
	b, _ := json.Marshal(creditPayload)

	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return &core.Secret{
				Name:  "amex",
				Type:  secrets.TypeCredit,
				Salt:  []byte("stub-salt-16byte"),
				Nonce: []byte("stub-nonce-12b"),
			}, b, nil
		},
	}
	uc := app.NewUpdateSecretUC(repo, &testutil.MockEncryptor{})

	newPayload := &secrets.CreditSecret{
		Number:          "4532123456789012",
		SecurityCode:    []byte("456"),
		ExpirationMonth: 12,
		ExpirationYear:  2030,
		Pin:             []byte("9999"),
	}
	req := app.UpdateSecretRequest{
		Name:           "amex",
		MasterPassword: []byte("master"),
		NewPayload:     newPayload,
	}
	if err := uc.Execute(context.Background(), req); err != nil {
		t.Fatalf("Execute() credit error = %v", err)
	}
}

func TestUpdateSecretUC_Execute_TypeMismatch(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: makeGetByNameFn(secrets.TypeAccount, "alice", "old"),
	}
	uc := app.NewUpdateSecretUC(repo, &testutil.MockEncryptor{})

	req := app.UpdateSecretRequest{
		Name:           "github",
		MasterPassword: []byte("master"),
		NewPayload: &secrets.CreditSecret{
			Number:          "4532123456789012",
			SecurityCode:    []byte("123"),
			ExpirationMonth: 6,
			ExpirationYear:  2028,
			Pin:             []byte("1234"),
		},
	}
	err := uc.Execute(context.Background(), req)
	if !core.IsDomainError(err) {
		t.Errorf("Execute() with type mismatch should return DomainError, got %v", err)
	}
}

func TestUpdateSecretUC_Execute_NotFound(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return nil, nil, core.ErrSecretNotFound
		},
	}
	uc := app.NewUpdateSecretUC(repo, &testutil.MockEncryptor{})

	req := app.UpdateSecretRequest{
		Name:           "missing",
		MasterPassword: []byte("master"),
		NewPayload:     &secrets.AccountSecret{Username: "u", Password: []byte("p")},
	}
	err := uc.Execute(context.Background(), req)
	if !core.IsNotFound(err) {
		t.Errorf("Execute() error = %v, want ErrSecretNotFound", err)
	}
}

func TestUpdateSecretUC_Execute_WrongMasterPassword(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: makeGetByNameFn(secrets.TypeAccount, "alice", "old"),
	}
	enc := &testutil.MockEncryptor{
		DecryptFn: func(_ context.Context, _ []byte, _ []byte, _ []byte, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	uc := app.NewUpdateSecretUC(repo, enc)

	req := app.UpdateSecretRequest{
		Name:           "github",
		MasterPassword: []byte("wrong"),
		NewPayload:     &secrets.AccountSecret{Username: "u", Password: []byte("p")},
	}
	err := uc.Execute(context.Background(), req)
	if err != core.ErrDecryptionFailed {
		t.Errorf("Execute() error = %v, want ErrDecryptionFailed", err)
	}
}

func TestUpdateSecretUC_Execute_EmptyName(t *testing.T) {
	uc := app.NewUpdateSecretUC(&testutil.MockRepository{}, &testutil.MockEncryptor{})
	req := app.UpdateSecretRequest{
		MasterPassword: []byte("master"),
		NewPayload:     &secrets.AccountSecret{Username: "u", Password: []byte("p")},
	}
	if err := uc.Execute(context.Background(), req); !core.IsDomainError(err) {
		t.Errorf("Execute() with empty name should return DomainError, got %v", err)
	}
}

func TestUpdateSecretUC_Execute_NilPayload(t *testing.T) {
	uc := app.NewUpdateSecretUC(&testutil.MockRepository{}, &testutil.MockEncryptor{})
	req := app.UpdateSecretRequest{Name: "svc", MasterPassword: []byte("master")}
	if err := uc.Execute(context.Background(), req); !core.IsDomainError(err) {
		t.Errorf("Execute() with nil payload should return DomainError, got %v", err)
	}
}

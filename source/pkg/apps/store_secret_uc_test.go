package apps_test

import (
	"context"
	"errors"
	"testing"

	"vextpss/source/core"
	"vextpss/source/pkg/apps"
	"vextpss/source/tests/mocks"
)

func TestStoreSecretUC_Execute_Success(t *testing.T) {
	repo := &mocks.MockRepository{}
	enc := &mocks.MockEncryptor{}
	uc := apps.NewStoreSecretUC(repo, enc)

	req := apps.StoreSecretRequest{
		Name:            "github",
		Type:            "account",
		MasterPassword:  []byte("masterpass"),
		AccountUsername: "alice",
		AccountPassword: []byte("s3cr3t"),
	}

	if err := uc.Execute(context.Background(), req); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
}

func TestStoreSecretUC_Execute_SavedSecretHasCorrectName(t *testing.T) {
	var savedName string
	repo := &mocks.MockRepository{
		SaveFn: func(_ context.Context, s *core.Secret, _ []byte) error {
			savedName = s.Name
			return nil
		},
	}
	uc := apps.NewStoreSecretUC(repo, &mocks.MockEncryptor{})

	req := apps.StoreSecretRequest{
		Name:            "my-service",
		Type:            "account",
		MasterPassword:  []byte("master"),
		AccountUsername: "user",
		AccountPassword: []byte("pass"),
	}
	uc.Execute(context.Background(), req)

	if savedName != "my-service" {
		t.Errorf("saved secret name = %q, want %q", savedName, "my-service")
	}
}

func TestStoreSecretUC_Execute_EmptyName(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := apps.StoreSecretRequest{
		Type:            "account",
		MasterPassword:  []byte("master"),
		AccountUsername: "user",
		AccountPassword: []byte("pass"),
	}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with empty name should return an error")
	}
	if !core.IsDomainError(err) {
		t.Errorf("error should be a DomainError, got %T: %v", err, err)
	}
}

func TestStoreSecretUC_Execute_UnsupportedType(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := apps.StoreSecretRequest{
		Name:           "svc",
		Type:           "card",
		MasterPassword: []byte("master"),
	}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with unsupported type should return an error")
	}
	if !core.IsDomainError(err) {
		t.Errorf("error should be a DomainError, got %T", err)
	}
}

func TestStoreSecretUC_Execute_EmptyMasterPassword(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := apps.StoreSecretRequest{
		Name:            "svc",
		Type:            "account",
		MasterPassword:  []byte{},
		AccountUsername: "user",
		AccountPassword: []byte("pass"),
	}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with empty master password should return an error")
	}
	if !core.IsDomainError(err) {
		t.Errorf("error should be a DomainError, got %T", err)
	}
}

func TestStoreSecretUC_Execute_EmptyUsername(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := apps.StoreSecretRequest{
		Name:            "svc",
		Type:            "account",
		MasterPassword:  []byte("master"),
		AccountUsername: "",
		AccountPassword: []byte("pass"),
	}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with empty username should return an error")
	}
	if !core.IsDomainError(err) {
		t.Errorf("error should be a DomainError, got %T", err)
	}
}

func TestStoreSecretUC_Execute_EmptyAccountPassword(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := apps.StoreSecretRequest{
		Name:            "svc",
		Type:            "account",
		MasterPassword:  []byte("master"),
		AccountUsername: "user",
		AccountPassword: []byte{},
	}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with empty account password should return an error")
	}
	if !core.IsDomainError(err) {
		t.Errorf("error should be a DomainError, got %T", err)
	}
}

func TestStoreSecretUC_Execute_AlreadyExists(t *testing.T) {
	repo := &mocks.MockRepository{
		SaveFn: func(_ context.Context, _ *core.Secret, _ []byte) error {
			return core.ErrAlreadyExists
		},
	}
	uc := apps.NewStoreSecretUC(repo, &mocks.MockEncryptor{})
	req := apps.StoreSecretRequest{
		Name:            "svc",
		Type:            "account",
		MasterPassword:  []byte("master"),
		AccountUsername: "user",
		AccountPassword: []byte("pass"),
	}
	err := uc.Execute(context.Background(), req)
	if !core.IsAlreadyExists(err) {
		t.Errorf("Execute() error = %v, want ErrAlreadyExists", err)
	}
}

func TestStoreSecretUC_Execute_EncryptionError(t *testing.T) {
	encErr := errors.New("crypto device failure")
	enc := &mocks.MockEncryptor{
		EncryptFn: func(_ context.Context, _, _ []byte) ([]byte, []byte, []byte, error) {
			return nil, nil, nil, encErr
		},
	}
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, enc)
	req := apps.StoreSecretRequest{
		Name:            "svc",
		Type:            "account",
		MasterPassword:  []byte("master"),
		AccountUsername: "user",
		AccountPassword: []byte("pass"),
	}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with encryption error should return an error")
	}
}

func TestStoreSecretUC_Execute_RepoTechnicalError(t *testing.T) {
	techErr := errors.New("db connection lost")
	repo := &mocks.MockRepository{
		SaveFn: func(_ context.Context, _ *core.Secret, _ []byte) error {
			return techErr
		},
	}
	uc := apps.NewStoreSecretUC(repo, &mocks.MockEncryptor{})
	req := apps.StoreSecretRequest{
		Name:            "svc",
		Type:            "account",
		MasterPassword:  []byte("master"),
		AccountUsername: "user",
		AccountPassword: []byte("pass"),
	}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with repo error should return an error")
	}
}

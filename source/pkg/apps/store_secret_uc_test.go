package apps_test

import (
	"context"
	"errors"
	"testing"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/pkg/apps"
	"vextpss/source/tests/mocks"
)

func validAccountRequest(name string) apps.StoreSecretRequest {
	return apps.StoreSecretRequest{
		Name:           name,
		MasterPassword: []byte("masterpass"),
		Payload: &secrets.AccountSecret{
			Username: "alice",
			Password: []byte("s3cr3t"),
		},
	}
}

func validCreditRequest(name string) apps.StoreSecretRequest {
	return apps.StoreSecretRequest{
		Name:           name,
		MasterPassword: []byte("masterpass"),
		Payload: &secrets.CreditSecret{
			Number:          "4532123456789012",
			SecurityCode:    []byte("123"),
			ExpirationMonth: 6,
			ExpirationYear:  2028,
			Pin:             []byte("1234"),
		},
	}
}

func TestStoreSecretUC_Execute_AccountSuccess(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	if err := uc.Execute(context.Background(), validAccountRequest("github")); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
}

func TestStoreSecretUC_Execute_CreditSuccess(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	if err := uc.Execute(context.Background(), validCreditRequest("visa-debit")); err != nil {
		t.Fatalf("Execute() credit error = %v, want nil", err)
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
	uc.Execute(context.Background(), validAccountRequest("my-service"))

	if savedName != "my-service" {
		t.Errorf("saved secret name = %q, want %q", savedName, "my-service")
	}
}

func TestStoreSecretUC_Execute_SavedCreditHasCorrectType(t *testing.T) {
	var savedType string
	repo := &mocks.MockRepository{
		SaveFn: func(_ context.Context, s *core.Secret, _ []byte) error {
			savedType = s.Type
			return nil
		},
	}
	uc := apps.NewStoreSecretUC(repo, &mocks.MockEncryptor{})
	uc.Execute(context.Background(), validCreditRequest("visa"))

	if savedType != "credit" {
		t.Errorf("saved secret type = %q, want %q", savedType, "credit")
	}
}

func TestStoreSecretUC_Execute_EmptyName(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := validAccountRequest("")
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with empty name should return an error")
	}
	if !core.IsDomainError(err) {
		t.Errorf("error should be a DomainError, got %T: %v", err, err)
	}
}

func TestStoreSecretUC_Execute_NilPayload(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := apps.StoreSecretRequest{
		Name:           "svc",
		MasterPassword: []byte("master"),
		Payload:        nil,
	}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with nil payload should return an error")
	}
	if !core.IsDomainError(err) {
		t.Errorf("error should be a DomainError, got %T", err)
	}
}

func TestStoreSecretUC_Execute_EmptyMasterPassword(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := validAccountRequest("svc")
	req.MasterPassword = []byte{}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with empty master password should return an error")
	}
	if !core.IsDomainError(err) {
		t.Errorf("error should be a DomainError, got %T", err)
	}
}

func TestStoreSecretUC_Execute_PayloadValidationFails(t *testing.T) {
	uc := apps.NewStoreSecretUC(&mocks.MockRepository{}, &mocks.MockEncryptor{})
	req := apps.StoreSecretRequest{
		Name:           "svc",
		MasterPassword: []byte("master"),
		Payload: &secrets.AccountSecret{
			Username: "",
			Password: []byte("pass"),
		},
	}
	err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Execute() with invalid payload should return an error")
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
	err := uc.Execute(context.Background(), validAccountRequest("svc"))
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
	err := uc.Execute(context.Background(), validAccountRequest("svc"))
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
	err := uc.Execute(context.Background(), validAccountRequest("svc"))
	if err == nil {
		t.Fatal("Execute() with repo error should return an error")
	}
}

package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/app"
	"vextpss/source/testutil"
)

func makeAccountPayload(t *testing.T, username string, password []byte) []byte {
	t.Helper()
	b, err := json.Marshal(&secrets.AccountSecret{Username: username, Password: password})
	if err != nil {
		t.Fatalf("failed to marshal AccountSecret: %v", err)
	}
	return b
}

func TestRetrieveSecretUC_Execute_Success(t *testing.T) {
	payload := makeAccountPayload(t, "alice", []byte("s3cr3t"))

	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, name string) (*core.Secret, []byte, error) {
			return &core.Secret{
				Name:      name,
				Type:      "account",
				Salt:      []byte("salt"),
				Nonce:     []byte("nonce"),
				CreatedAt: time.Now(),
			}, payload, nil
		},
	}
	// Transparent decryptor: returns ciphertext as plaintext.
	enc := &testutil.MockEncryptor{}
	uc := app.NewRetrieveSecretUC(repo, enc)

	req := app.RetrieveSecretRequest{Name: "github", MasterPassword: []byte("master")}
	resp, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.Name != "github" {
		t.Errorf("resp.Name = %q, want %q", resp.Name, "github")
	}
	if resp.Type != "account" {
		t.Errorf("resp.Type = %q, want %q", resp.Type, "account")
	}
	account, ok := resp.Payload.(*secrets.AccountSecret)
	if !ok {
		t.Fatalf("payload type = %T, want *secrets.AccountSecret", resp.Payload)
	}
	if account.Username != "alice" {
		t.Errorf("account.Username = %q, want %q", account.Username, "alice")
	}
}

func TestRetrieveSecretUC_Execute_NotFound(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return nil, nil, core.ErrSecretNotFound
		},
	}
	uc := app.NewRetrieveSecretUC(repo, &testutil.MockEncryptor{})

	_, err := uc.Execute(context.Background(), app.RetrieveSecretRequest{
		Name:           "missing",
		MasterPassword: []byte("master"),
	})
	if !core.IsNotFound(err) {
		t.Errorf("Execute() error = %v, want ErrSecretNotFound", err)
	}
}

func TestRetrieveSecretUC_Execute_DecryptionFailed(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, name string) (*core.Secret, []byte, error) {
			return &core.Secret{Name: name, Type: "account"}, []byte("encrypted"), nil
		},
	}
	enc := &testutil.MockEncryptor{
		DecryptFn: func(_ context.Context, _, _, _, _ []byte) ([]byte, error) {
			return nil, core.ErrDecryptionFailed
		},
	}
	uc := app.NewRetrieveSecretUC(repo, enc)

	_, err := uc.Execute(context.Background(), app.RetrieveSecretRequest{
		Name:           "svc",
		MasterPassword: []byte("wrong"),
	})
	if err == nil {
		t.Fatal("Execute() with bad password should return an error")
	}
	if err.Error() != core.ErrDecryptionFailed.Error() {
		t.Errorf("Execute() error = %q, want %q", err.Error(), core.ErrDecryptionFailed.Error())
	}
}

func TestRetrieveSecretUC_Execute_RepoError(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, _ string) (*core.Secret, []byte, error) {
			return nil, nil, errors.New("db down")
		},
	}
	uc := app.NewRetrieveSecretUC(repo, &testutil.MockEncryptor{})

	_, err := uc.Execute(context.Background(), app.RetrieveSecretRequest{
		Name:           "svc",
		MasterPassword: []byte("master"),
	})
	if err == nil {
		t.Fatal("Execute() with repo error should return an error")
	}
}

func TestRetrieveSecretUC_Execute_CreditSuccess(t *testing.T) {
	creditPayload := &secrets.CreditSecret{
		Number:          "4532123456789012",
		SecurityCode:    []byte("123"),
		ExpirationMonth: 6,
		ExpirationYear:  2028,
		Pin:             []byte("1234"),
	}
	b, _ := json.Marshal(creditPayload)

	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, name string) (*core.Secret, []byte, error) {
			return &core.Secret{Name: name, Type: "credit", CreatedAt: time.Now()}, b, nil
		},
	}
	uc := app.NewRetrieveSecretUC(repo, &testutil.MockEncryptor{})

	resp, err := uc.Execute(context.Background(), app.RetrieveSecretRequest{
		Name:           "visa-debit",
		MasterPassword: []byte("master"),
	})
	if err != nil {
		t.Fatalf("Execute() credit error = %v", err)
	}
	if resp.Type != "credit" {
		t.Errorf("resp.Type = %q, want %q", resp.Type, "credit")
	}
	credit, ok := resp.Payload.(*secrets.CreditSecret)
	if !ok {
		t.Fatalf("payload type = %T, want *secrets.CreditSecret", resp.Payload)
	}
	if credit.Number != "4532123456789012" {
		t.Errorf("credit.Number = %q, want %q", credit.Number, "4532123456789012")
	}
}

func TestRetrieveSecretUC_Execute_UnknownSecretType(t *testing.T) {
	repo := &testutil.MockRepository{
		GetByNameFn: func(_ context.Context, name string) (*core.Secret, []byte, error) {
			return &core.Secret{Name: name, Type: "note"}, []byte(`{}`), nil
		},
	}
	uc := app.NewRetrieveSecretUC(repo, &testutil.MockEncryptor{})

	_, err := uc.Execute(context.Background(), app.RetrieveSecretRequest{
		Name:           "svc",
		MasterPassword: []byte("master"),
	})
	if err == nil {
		t.Fatal("Execute() with unknown secret type should return an error")
	}
}

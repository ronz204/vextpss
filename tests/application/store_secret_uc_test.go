package application_test

import (
	"context"
	"testing"
	"vextpss/source/pkg/application"
	"vextpss/source/pkg/domain"
)

// --- Mock implementations ---

type mockRepo struct {
	saved          *domain.Secret
	savedPayload   []byte
	getByNameFn    func(name string) (*domain.Secret, []byte, error)
	deleteFn       func(name string) error
	listAllSecrets []domain.Secret
}

func (m *mockRepo) Save(_ context.Context, s *domain.Secret, p []byte) error {
	m.saved = s
	m.savedPayload = p
	return nil
}

func (m *mockRepo) GetByName(_ context.Context, name string) (*domain.Secret, []byte, error) {
	if m.getByNameFn != nil {
		return m.getByNameFn(name)
	}
	return nil, nil, domain.ErrSecretNotFound
}

func (m *mockRepo) ListAll(_ context.Context) ([]domain.Secret, error) {
	return m.listAllSecrets, nil
}

func (m *mockRepo) Delete(_ context.Context, name string) error {
	if m.deleteFn != nil {
		return m.deleteFn(name)
	}
	return nil
}

type mockEncryptor struct {
	encryptFn func(plaintext, masterPwd []byte) ([]byte, []byte, []byte, error)
	decryptFn func(masterPwd, salt, nonce, ciphertext []byte) ([]byte, error)
}

func (m *mockEncryptor) Encrypt(_ context.Context, plaintext, masterPwd []byte) ([]byte, []byte, []byte, error) {
	if m.encryptFn != nil {
		return m.encryptFn(plaintext, masterPwd)
	}
	return []byte("salt"), []byte("nonce"), plaintext, nil // identity for simplicity
}

func (m *mockEncryptor) Decrypt(_ context.Context, masterPwd, salt, nonce, ciphertext []byte) ([]byte, error) {
	if m.decryptFn != nil {
		return m.decryptFn(masterPwd, salt, nonce, ciphertext)
	}
	return ciphertext, nil
}

// --- Tests ---

func TestStoreSecretUC_Execute_HappyPath(t *testing.T) {
	repo := &mockRepo{}
	enc := &mockEncryptor{}
	uc := application.NewStoreSecretUC(repo, enc)

	err := uc.Execute(context.Background(), application.StoreSecretRequest{
		Name:            "github",
		Type:            "account",
		MasterPassword:  []byte("master"),
		AccountUsername: "alice",
		AccountPassword: []byte("s3cr3t"),
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected repo.Save to be called")
	}
	if repo.saved.Name != "github" {
		t.Errorf("saved name = %q, want %q", repo.saved.Name, "github")
	}
}

func TestStoreSecretUC_Execute_EmptyNameReturnsError(t *testing.T) {
	uc := application.NewStoreSecretUC(&mockRepo{}, &mockEncryptor{})

	err := uc.Execute(context.Background(), application.StoreSecretRequest{
		Name:           "",
		Type:           "account",
		MasterPassword: []byte("master"),
	})

	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !domain.IsDomainError(err) {
		t.Errorf("expected DomainError, got %T: %v", err, err)
	}
}

func TestStoreSecretUC_Execute_UnsupportedTypeReturnsError(t *testing.T) {
	uc := application.NewStoreSecretUC(&mockRepo{}, &mockEncryptor{})

	err := uc.Execute(context.Background(), application.StoreSecretRequest{
		Name:           "k",
		Type:           "ssh_key",
		MasterPassword: []byte("master"),
	})

	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestDeleteSecretUC_Execute_NotFound(t *testing.T) {
	repo := &mockRepo{
		deleteFn: func(name string) error { return domain.ErrSecretNotFound },
	}
	uc := application.NewDeleteSecretUC(repo)

	err := uc.Execute(context.Background(), "nonexistent")
	if !domain.IsNotFound(err) {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestListSecretsUC_Execute_ReturnsAll(t *testing.T) {
	repo := &mockRepo{
		listAllSecrets: []domain.Secret{
			{Name: "a"}, {Name: "b"},
		},
	}
	uc := application.NewListSecretsUC(repo)

	secrets, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(secrets))
	}
}

package mocks

import "context"

// MockEncryptor is a configurable test double for tokens.Encryptor.
// Default behavior: Encrypt returns plaintext as ciphertext (with stub salt/nonce).
// Default behavior: Decrypt returns ciphertext as plaintext (transparent round-trip).
type MockEncryptor struct {
	EncryptFn func(ctx context.Context, plaintext, masterPassword []byte) (salt, nonce, ciphertext []byte, err error)
	DecryptFn func(ctx context.Context, masterPassword, salt, nonce, ciphertext []byte) ([]byte, error)
}

func (m *MockEncryptor) Encrypt(
	ctx context.Context,
	plaintext, masterPassword []byte,
) (salt, nonce, ciphertext []byte, err error) {
	if m.EncryptFn != nil {
		return m.EncryptFn(ctx, plaintext, masterPassword)
	}
	return []byte("stub-salt-16byte"), []byte("stub-nonce-12b"), plaintext, nil
}

func (m *MockEncryptor) Decrypt(
	ctx context.Context,
	masterPassword, salt, nonce, ciphertext []byte,
) ([]byte, error) {
	if m.DecryptFn != nil {
		return m.DecryptFn(ctx, masterPassword, salt, nonce, ciphertext)
	}
	return ciphertext, nil
}

package testutil

import "context"

// MockEncryptor is a configurable test double for crypto.Encryptor.
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
	// Return a copy so callers that zero plaintext do not corrupt the returned ciphertext.
	result := make([]byte, len(plaintext))
	copy(result, plaintext)
	return []byte("stub-salt-16byte"), []byte("stub-nonce-12b"), result, nil
}

func (m *MockEncryptor) Decrypt(
	ctx context.Context,
	masterPassword, salt, nonce, ciphertext []byte,
) ([]byte, error) {
	if m.DecryptFn != nil {
		return m.DecryptFn(ctx, masterPassword, salt, nonce, ciphertext)
	}
	// Return a copy so callers that zero the result do not corrupt the ciphertext in storage.
	result := make([]byte, len(ciphertext))
	copy(result, ciphertext)
	return result, nil
}

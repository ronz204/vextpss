package core

import "context"

// Encryptor is the cryptographic contract consumed by use cases.
// Implementations live in pkg/tokens/ — never in core or application.
type Encryptor interface {
	// Encrypt ciphers plaintext with masterPassword.
	// Returns a unique salt, nonce, and ciphertext to be stored alongside the record.
	Encrypt(ctx context.Context, plaintext, masterPassword []byte) (salt, nonce, ciphertext []byte, err error)

	// Decrypt reverses Encrypt. Returns ErrDecryptionFailed on wrong password or tampered data.
	Decrypt(ctx context.Context, masterPassword, salt, nonce, ciphertext []byte) ([]byte, error)
}

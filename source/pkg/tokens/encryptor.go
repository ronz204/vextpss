package tokens

import "context"

// Encryptor defines the contract for symmetric encryption and decryption.
// Implementations live in infrastructure/crypto — never in domain or application.
type Encryptor interface {
	// Encrypt ciphers plaintext with masterPassword.
	// Returns salt, nonce, and ciphertext to be stored alongside the record.
	Encrypt(ctx context.Context, plaintext, masterPassword []byte) (salt, nonce, ciphertext []byte, err error)

	// Decrypt reverses Encrypt. Returns ErrDecryptionFailed on wrong password or tampered data.
	Decrypt(ctx context.Context, masterPassword, salt, nonce, ciphertext []byte) ([]byte, error)
}

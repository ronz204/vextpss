package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 2
	argonKeyLen  = 32
)

// ErrDecryptFailed is returned when AES-GCM authentication fails.
// It is intentionally generic to avoid leaking failure mode details.
var ErrDecryptFailed = errors.New("master password incorrect or data corrupted")

// DeriveKey derives a 32-byte encryption key from a password and salt using Argon2id.
// The salt must be unique per record (16 random bytes).
func DeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
}

// Encrypt encrypts plaintext using AES-256-GCM with the given key.
// It generates a fresh random 12-byte nonce for each call.
// Returns the nonce and ciphertext separately so the nonce can be stored alongside the encrypted data.
func Encrypt(key, plaintext []byte) (nonce []byte, ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the given key and nonce.
// Returns ErrDecryptFailed if the authentication tag check fails (wrong key or tampered data).
func Decrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrDecryptFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptFailed
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptFailed
	}

	return plaintext, nil
}

// GenerateSalt generates a cryptographically secure random 16-byte salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// Zero overwrites a byte slice with zeros to clear sensitive data from memory.
// This must be called immediately after use of any slice holding plaintext secrets or keys.
func Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

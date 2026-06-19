package cryptors

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"vextpss/source/shared/memory"
	"vextpss/source/shared/sentinel"

	"golang.org/x/crypto/argon2"
)

type AESGCMEncryptor struct {
	config AESGCMConfig
}

func NewAESGCMEncryptor(config AESGCMConfig) *AESGCMEncryptor {
	return &AESGCMEncryptor{config: config}
}

// ======================================
// Encrypt derives a key from password + a fresh salt, then encrypts plaintext with AES-256-GCM.
// ======================================
func (e *AESGCMEncryptor) Encrypt(_ context.Context, plaintext, password []byte) (salt, nonce, ciphertext []byte, err error) {
	salt, err = e.randomBytes(e.config.SaltLen)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := e.deriveKey(password, salt)
	defer memory.Cleaner(key)

	nonce, err = e.randomBytes(e.config.NonceLen)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	gcm, err := e.newGCM(key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return salt, nonce, ciphertext, nil
}

// ======================================
// Decrypt derives the same key from password + stored salt, then deciphers the payload.
// ======================================
func (e *AESGCMEncryptor) Decrypt(_ context.Context, password, salt, nonce, ciphertext []byte) ([]byte, error) {
	if len(salt) != e.config.SaltLen || len(nonce) != e.config.NonceLen {
		return nil, sentinel.ErrDecryptionFailed
	}

	key := e.deriveKey(password, salt)
	defer memory.Cleaner(key)

	gcm, err := e.newGCM(key)
	if err != nil {
		return nil, sentinel.ErrDecryptionFailed
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, sentinel.ErrDecryptionFailed
	}

	return plaintext, nil
}

func (e *AESGCMEncryptor) deriveKey(password, salt []byte) []byte {
	a := e.config.Argon
	return argon2.IDKey(password, salt, a.Time, a.Memory, a.Threads, a.KeyLen)
}

func (e *AESGCMEncryptor) newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (e *AESGCMEncryptor) randomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

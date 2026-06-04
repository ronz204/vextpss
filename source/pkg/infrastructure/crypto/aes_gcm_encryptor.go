package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"vextpss/source/pkg/domain"
	"vextpss/source/pkg/shared"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024 // 64 MB
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	saltLen             = 16
	nonceLen            = 12
)

// AESGCMEncryptor implements domain.Encryptor using AES-256-GCM + Argon2id key derivation.
type AESGCMEncryptor struct{}

// NewAESGCMEncryptor returns an encryptor with production-grade Argon2id parameters.
func NewAESGCMEncryptor() *AESGCMEncryptor {
	return &AESGCMEncryptor{}
}

// Encrypt derives a key from masterPassword + a fresh salt, then encrypts plaintext with AES-256-GCM.
func (e *AESGCMEncryptor) Encrypt(
	_ context.Context,
	plaintext, masterPassword []byte,
) (salt, nonce, ciphertext []byte, err error) {
	salt, err = shared.GenerateSalt()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := e.deriveKey(masterPassword, salt)
	defer shared.Zero(key)

	nonce = make([]byte, nonceLen)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return salt, nonce, ciphertext, nil
}

// Decrypt derives the same key from masterPassword + stored salt, then deciphers the payload.
// Returns domain.ErrDecryptionFailed on wrong password or tampered data — never raw crypto errors.
func (e *AESGCMEncryptor) Decrypt(
	_ context.Context,
	masterPassword, salt, nonce, ciphertext []byte,
) ([]byte, error) {
	key := e.deriveKey(masterPassword, salt)
	defer shared.Zero(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, domain.ErrDecryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, domain.ErrDecryptionFailed
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, domain.ErrDecryptionFailed
	}

	return plaintext, nil
}

func (e *AESGCMEncryptor) deriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
}

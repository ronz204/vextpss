package crypto

import (
	ctx "context"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"vextpss/source/core"
	"vextpss/source/errors"
	"vextpss/source/shared/password"

	"golang.org/x/crypto/argon2"
)

// AESGCMEncryptor implements core.Encryptor using AES-GCM and Argon2id.
type AESGCMEncryptor struct {
	config AESGCMConfig
}

// NewAESGCMEncryptor returns an encryptor with the provided configuration.
func NewAESGCMEncryptor(config AESGCMConfig) *AESGCMEncryptor {
	return &AESGCMEncryptor{config: config}
}

// Encrypt derives a key from masterPassword + a fresh salt, then encrypts plaintext with AES-256-GCM.
func (e *AESGCMEncryptor) Encrypt(ctx ctx.Context, in core.EncInDto) (core.EncOutDto, error) {
	salt, err := e.randomBytes(e.config.SaltLen)
	if err != nil {
		return core.EncOutDto{}, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := e.deriveKey(in.Password, salt)
	defer password.Cleaner(key)

	nonce, err := e.randomBytes(e.config.NonceLen)
	if err != nil {
		return core.EncOutDto{}, fmt.Errorf("failed to generate nonce: %w", err)
	}

	gcm, err := e.newGCM(key)
	if err != nil {
		return core.EncOutDto{}, fmt.Errorf("failed to create cipher: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, in.Plaintext, nil)

	return core.EncOutDto{
		Salt:       salt,
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}, nil
}

// Decrypt derives the same key from masterPassword + stored salt, then deciphers the payload.
func (e *AESGCMEncryptor) Decrypt(ctx ctx.Context, in core.DecInDto) ([]byte, error) {
	if len(in.Salt) != e.config.SaltLen || len(in.Nonce) != e.config.NonceLen {
		return nil, errors.ErrDecryptionFailed
	}

	key := e.deriveKey(in.Password, in.Salt)
	defer password.Cleaner(key)

	gcm, err := e.newGCM(key)
	if err != nil {
		return nil, errors.ErrDecryptionFailed
	}

	plaintext, err := gcm.Open(nil, in.Nonce, in.Ciphertext, nil)
	if err != nil {
		return nil, errors.ErrDecryptionFailed
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

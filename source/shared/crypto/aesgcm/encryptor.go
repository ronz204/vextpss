package aesgcm

import (
	"context"
	"crypto/aes"
	"crypto/cipher"

	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
)

type AESGCMEncryptor struct {
	config AESGCMConfig
}

func NewAESGCMEncryptor(config AESGCMConfig) *AESGCMEncryptor {
	return &AESGCMEncryptor{config: config}
}

func (e *AESGCMEncryptor) Encrypt(_ context.Context, plaintext, password []byte) (*AESGCMPayload, error) {
	salt, err := randomBytes(e.config.SaltLen)
	if err != nil {
		return nil, err
	}

	nonce, err := randomBytes(e.config.NonceLen)
	if err != nil {
		return nil, err
	}

	key := deriveKey(password, salt, e.config.Argon)
	defer memory.Cleaner(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return &AESGCMPayload{
		Salt:       salt,
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}, nil
}

func (e *AESGCMEncryptor) Decrypt(_ context.Context, data *AESGCMPayload, password []byte) ([]byte, error) {
	key := deriveKey(password, data.Salt, e.config.Argon)
	defer memory.Cleaner(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, data.Nonce, data.Ciphertext, nil)
	if err != nil {
		return nil, secrets.ErrDecryptionFailed
	}

	return plaintext, nil
}

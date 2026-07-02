package aesgcm

import (
	"context"
	"fmt"

	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
)

type Encryptor struct {
	config Config
}

func New(config Config) *Encryptor {
	return &Encryptor{config: config}
}

func (e *Encryptor) Algorithm() string {
	return algorithmID
}

func (e *Encryptor) Encrypt(_ context.Context, plaintext, password []byte) (secrets.Encrypted, error) {
	salt, err := randomBytes(e.config.SaltLen)
	if err != nil {
		return secrets.Encrypted{}, err
	}

	nonce, err := randomBytes(e.config.NonceLen)
	if err != nil {
		return secrets.Encrypted{}, err
	}

	key := deriveKey(password, salt, e.config.Argon)
	defer memory.Cleaner(key)

	gcm, err := newGCM(key)
	if err != nil {
		return secrets.Encrypted{}, err
	}

	return secrets.Encrypted{
		Algorithm:  algorithmID,
		Ciphertext: gcm.Seal(nil, nonce, plaintext, nil),
		Metadata:   encodeMetadata(salt, nonce),
	}, nil
}

func (e *Encryptor) Decrypt(_ context.Context, payload secrets.Encrypted, password []byte) ([]byte, error) {
	if payload.Algorithm != algorithmID {
		return nil, fmt.Errorf("aesgcm: unsupported algorithm %q: %w", payload.Algorithm, secrets.ErrUnsupportedAlgorithm)
	}

	salt, nonce, err := decodeMetadata(payload.Metadata)
	if err != nil {
		return nil, err
	}

	key := deriveKey(password, salt, e.config.Argon)
	defer memory.Cleaner(key)

	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, payload.Ciphertext, nil)
	if err != nil {
		return nil, secrets.ErrDecryptionFailed
	}

	return plaintext, nil
}

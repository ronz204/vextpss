package cryptors

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"golang.org/x/crypto/argon2"
)

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

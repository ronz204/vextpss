package aesgcm

import (
	"crypto/rand"

	"golang.org/x/crypto/argon2"
)

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func deriveKey(password, salt []byte, config Argon2Config) []byte {
	return argon2.IDKey(password, salt, config.Time, config.Memory, config.Threads, config.KeyLen)
}

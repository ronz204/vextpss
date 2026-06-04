package shared

import (
	"crypto/rand"
	"io"
	"time"
)

// Zero overwrites a byte slice with zeros to clear sensitive data from memory.
// Call immediately after the slice is no longer needed.
func Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// GenerateSalt returns a cryptographically secure random 16-byte salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// Now returns the current UTC time. Centralised so tests can swap it out easily.
func Now() time.Time {
	return time.Now().UTC()
}

package models

import "time"

// SecretRecord is a 1:1 mapping of the `secrets` table.
// It contains no plaintext secret data — EncryptedPayload is always opaque bytes.
type SecretRecord struct {
	ID               int64
	Name             string
	Type             string
	Salt             []byte
	Nonce            []byte
	EncryptedPayload []byte
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

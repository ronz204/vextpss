package domain

import "time"

// SecretID is a value object representing the unique identity of a Secret.
type SecretID int64

// Secret is the core domain entity. It holds metadata and an opaque encrypted payload.
// It does not contain implementation details of DB or crypto.
type Secret struct {
	ID        SecretID
	Name      string
	Type      string
	Salt      []byte
	Nonce     []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SecretPayload is a polymorphic interface implemented by every secret type.
type SecretPayload interface {
	GetType() string
	Validate() error
}

package core

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
	// MergeFrom fills any zero-value fields from current. Called during update
	// so the handler stays type-agnostic — each type decides what merging means.
	MergeFrom(current SecretPayload)
}

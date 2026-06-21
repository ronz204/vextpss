package secrets

import "time"

type SecretID int64

type Secret struct {
	ID        SecretID
	Name      string
	Type      string
	Salt      []byte
	Nonce     []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Credential struct {
	Secret    Secret
	Encrypted []byte
}

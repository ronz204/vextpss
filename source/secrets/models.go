package secrets

import "time"

type Secret struct {
	ID        int64
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

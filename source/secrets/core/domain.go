package core

import (
	"time"
)

type Secret struct {
	ID        int64
	Name      string
	Type      string
	Encrypted Encrypted
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Payload interface {
	Type() string
	Display()
	Validate() error
}

const (
	TypeAccount = "account"
	TypeFinance = "finance"
)

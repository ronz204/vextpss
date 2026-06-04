package domain

import "errors"

// DomainError represents a business-level error (not a technical/infrastructure error).
type DomainError struct {
	message string
}

func NewDomainError(msg string) DomainError {
	return DomainError{message: msg}
}

func (e DomainError) Error() string {
	return e.message
}

// Canonical domain errors — infrastructure layers must map their own errors to these.
var (
	ErrSecretNotFound    = NewDomainError("secret not found")
	ErrAlreadyExists     = NewDomainError("secret already exists")
	ErrInvalidCredential = NewDomainError("invalid credential format")
	ErrDecryptionFailed  = NewDomainError("master password incorrect or data corrupted")
	ErrInvalidInput      = NewDomainError("invalid input")
)

// IsDomainError reports whether err is a DomainError.
func IsDomainError(err error) bool {
	var de DomainError
	return errors.As(err, &de)
}

// IsNotFound reports whether err signals a missing secret.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrSecretNotFound)
}

// IsAlreadyExists reports whether err signals a duplicate name.
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

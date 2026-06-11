package excepts

import "errors"

// === Represents a business-level error.
type DomainError struct {
	message string
}

func NewDomainError(msg string) DomainError {
	return DomainError{message: msg}
}

func (e DomainError) Error() string {
	return e.message
}

// === Canonical domain errors.
var (
	ErrSecretNotFound    = NewDomainError("secret not found")
	ErrAlreadyExists     = NewDomainError("secret already exists")
	ErrInvalidCredential = NewDomainError("invalid credential format")
	ErrDecryptionFailed  = NewDomainError("master password incorrect or data corrupted")
	ErrInvalidInput      = NewDomainError("invalid input")
)

// === Helper functions to check error types.
func IsDomainError(err error) bool {
	var de DomainError
	return errors.As(err, &de)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrSecretNotFound)
}

func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

package sentinel

// Sentinel is an immutable typed error constant.
// Being a named string type it satisfies the error interface and supports == comparisons.
type Sentinel string

func (s Sentinel) Error() string { return string(s) }

// == Domain sentinels ================

const (
	ErrSecretNotFound    Sentinel = "secret not found"
	ErrAlreadyExists     Sentinel = "secret already exists"
	ErrInvalidCredential Sentinel = "invalid credential format"
	ErrDecryptionFailed  Sentinel = "master password incorrect or data corrupted"
	ErrInvalidInput      Sentinel = "invalid input"
)

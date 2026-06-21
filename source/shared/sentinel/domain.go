package sentinel

type Sentinel string

func (s Sentinel) Error() string { return string(s) }

const (
	ErrSecretNotFound    Sentinel = "secret not found"
	ErrAlreadyExists     Sentinel = "secret already exists"
	ErrInvalidCredential Sentinel = "invalid credential format"
	ErrDecryptionFailed  Sentinel = "master password incorrect or data corrupted"
	ErrInvalidInput      Sentinel = "invalid input"
)

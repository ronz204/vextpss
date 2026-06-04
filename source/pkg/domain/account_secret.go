package domain

// AccountSecret is the payload for secrets of type "account".
// Password is []byte — never string — so it can be zeroed from memory after use.
type AccountSecret struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

func (a *AccountSecret) GetType() string { return "account" }

func (a *AccountSecret) Validate() error {
	if a.Username == "" {
		return NewDomainError("username is required")
	}
	if len(a.Password) == 0 {
		return NewDomainError("password is required")
	}
	if len(a.Username) > 255 {
		return NewDomainError("username exceeds maximum length")
	}
	return nil
}

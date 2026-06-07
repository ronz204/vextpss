package secrets

import "vextpss/source/core"

// AccountSecret is the payload for secrets of type "account".
// Password is []byte — never string — so it can be zeroed from memory after use.
type AccountSecret struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

func (a *AccountSecret) GetType() string { return TypeAccount }

func (a *AccountSecret) MergeFrom(current core.SecretPayload) {
	cur, ok := current.(*AccountSecret)
	if !ok || cur == nil {
		return
	}
	if a.Username == "" {
		a.Username = cur.Username
	}
}

func (a *AccountSecret) Validate() error {
	if a.Username == "" {
		return core.NewDomainError("username is required")
	}
	if len(a.Password) == 0 {
		return core.NewDomainError("password is required")
	}
	if len(a.Username) > 255 {
		return core.NewDomainError("username exceeds maximum length")
	}
	return nil
}

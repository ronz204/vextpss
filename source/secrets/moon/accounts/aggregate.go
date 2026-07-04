package accounts

import (
	"fmt"

	"vextpss/source/secrets/core"
)

var _ core.Payload = Account{}

type Account struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

func NewAccount(username string, password []byte) (Account, error) {
	account := Account{Username: username, Password: password}
	return account, account.Validate()
}

func (a Account) Type() string {
	return core.TypeAccount
}

func (a Account) Display() {
	fmt.Printf("  %-16s %s\n", "Username", a.Username)
	fmt.Printf("  %-16s %s\n", "Password", string(a.Password))
}

func (a Account) Validate() error {
	if a.Username == "" {
		return ErrUsernameRequired
	}
	if a.Password == nil {
		return ErrPasswordRequired
	}
	return nil
}

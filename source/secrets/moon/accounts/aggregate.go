package accounts

import "vextpss/source/secrets/core"

type Account struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

func NewAccount(username string, password []byte) (Account, error) {
	account := Account{Username: username, Password: password}
	return account, account.Validate()
}

func (a Account) Display() string {
	return core.TypeAccount
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

package secrets

import "fmt"

type Account struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

func NewAccount(username string, password []byte) (Account, error) {
	account := Account{Username: username, Password: password}
	return account, account.Validate()
}

func (a Account) Display() string {
	return TypeAccount
}

func (a Account) Validate() error {
	if a.Username == "" {
		return fmt.Errorf("username is required")
	}
	if a.Password == nil {
		return fmt.Errorf("password is required")
	}
	return nil
}

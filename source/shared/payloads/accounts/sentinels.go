package accounts

import "errors"

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrPasswordRequired = errors.New("password is required")
)

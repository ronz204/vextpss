package finances

import "errors"

var (
	ErrCardNumberRequired         = errors.New("card number is required")
	ErrCardPinRequired            = errors.New("card pin is required")
	ErrCardSecurityCodeRequired   = errors.New("card security code is required")
	ErrCardExpirationMonthInvalid = errors.New("card expiration month must be between 1 and 12")
	ErrCardExpirationYearRequired = errors.New("card expiration year is required")

	ErrMobileUsernameRequired   = errors.New("mobile username is required")
	ErrMobilePasswordRequired   = errors.New("mobile password is required")
	ErrMobileVirtualKeyRequired = errors.New("mobile virtual key is required")
	ErrMobileCellphoneRequired  = errors.New("mobile cellphone is required")
)

package finances

type Card struct {
	Pin             []byte `json:"pin"`
	Number          string `json:"number"`
	SecurityCode    []byte `json:"security_code"`
	ExpirationMonth int    `json:"expiration_month"`
	ExpirationYear  int    `json:"expiration_year"`
}

func NewCard(number string, pin, code []byte, expMonth, expYear int) (Card, error) {
	c := Card{Number: number, Pin: pin, SecurityCode: code, ExpirationMonth: expMonth, ExpirationYear: expYear}
	return c, c.Validate()
}

func (c Card) Validate() error {
	if c.Number == "" {
		return ErrCardNumberRequired
	}
	if c.Pin == nil {
		return ErrCardPinRequired
	}
	if c.SecurityCode == nil {
		return ErrCardSecurityCodeRequired
	}
	if c.ExpirationMonth < 1 || c.ExpirationMonth > 12 {
		return ErrCardExpirationMonthInvalid
	}
	if c.ExpirationYear < 1 {
		return ErrCardExpirationYearRequired
	}
	return nil
}

type Mobile struct {
	Username   string `json:"username"`
	Password   []byte `json:"password"`
	VirtualKey []byte `json:"virtual_key"`
	Cellphone  string `json:"cellphone"`
}

func NewMobile(username string, password, virtualKey []byte, cellphone string) (Mobile, error) {
	m := Mobile{Username: username, Password: password, VirtualKey: virtualKey, Cellphone: cellphone}
	return m, m.Validate()
}

func (m Mobile) Validate() error {
	if m.Username == "" {
		return ErrMobileUsernameRequired
	}
	if m.Password == nil {
		return ErrMobilePasswordRequired
	}
	if m.VirtualKey == nil {
		return ErrMobileVirtualKeyRequired
	}
	if m.Cellphone == "" {
		return ErrMobileCellphoneRequired
	}
	return nil
}

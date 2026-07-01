package secrets

import "fmt"

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
		return fmt.Errorf("card number is required")
	}
	if c.Pin == nil {
		return fmt.Errorf("card pin is required")
	}
	if c.SecurityCode == nil {
		return fmt.Errorf("card security code is required")
	}
	if c.ExpirationMonth < 1 || c.ExpirationMonth > 12 {
		return fmt.Errorf("card expiration month must be between 1 and 12")
	}
	if c.ExpirationYear < 1 {
		return fmt.Errorf("card expiration year is required")
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
		return fmt.Errorf("mobile username is required")
	}
	if m.Password == nil {
		return fmt.Errorf("mobile password is required")
	}
	if m.VirtualKey == nil {
		return fmt.Errorf("mobile virtual key is required")
	}
	if m.Cellphone == "" {
		return fmt.Errorf("mobile cellphone is required")
	}
	return nil
}

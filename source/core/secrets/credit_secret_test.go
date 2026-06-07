package secrets_test

import (
	"testing"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
)

func validCreditSecret() *secrets.CreditSecret {
	return &secrets.CreditSecret{
		Number:          "4532123456789012",
		SecurityCode:    []byte("123"),
		ExpirationMonth: 6,
		ExpirationYear:  2028,
		Pin:             []byte("1234"),
	}
}

func TestCreditSecret_GetType(t *testing.T) {
	c := &secrets.CreditSecret{}
	if got := c.GetType(); got != "credit" {
		t.Errorf("GetType() = %q, want %q", got, "credit")
	}
}

func TestCreditSecret_Validate_Success(t *testing.T) {
	if err := validCreditSecret().Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestCreditSecret_Validate_WithOptionalFields(t *testing.T) {
	c := validCreditSecret()
	c.BankUsername = "user@bank.com"
	c.BankPassword = []byte("bankpass")
	c.BankVirtualKey = []byte("vk123")
	c.BankCellphone = "+57 300 123 4567"
	c.CountryCode = "CO"
	if err := c.Validate(); err != nil {
		t.Errorf("Validate() with optional fields error = %v, want nil", err)
	}
}

func TestCreditSecret_Validate_MissingNumber(t *testing.T) {
	c := validCreditSecret()
	c.Number = ""
	err := c.Validate()
	if err == nil {
		t.Fatal("Validate() should fail with empty card number")
	}
	if !core.IsDomainError(err) {
		t.Errorf("expected DomainError, got %T", err)
	}
}

func TestCreditSecret_Validate_MissingSecurityCode(t *testing.T) {
	c := validCreditSecret()
	c.SecurityCode = nil
	err := c.Validate()
	if err == nil {
		t.Fatal("Validate() should fail with empty security code")
	}
	if !core.IsDomainError(err) {
		t.Errorf("expected DomainError, got %T", err)
	}
}

func TestCreditSecret_Validate_InvalidMonth(t *testing.T) {
	cases := []int{0, 13, -1, 100}
	for _, month := range cases {
		c := validCreditSecret()
		c.ExpirationMonth = month
		if err := c.Validate(); err == nil {
			t.Errorf("Validate() with month=%d should fail", month)
		}
	}
}

func TestCreditSecret_Validate_InvalidYear(t *testing.T) {
	cases := []int{1999, 2101, 0}
	for _, year := range cases {
		c := validCreditSecret()
		c.ExpirationYear = year
		if err := c.Validate(); err == nil {
			t.Errorf("Validate() with year=%d should fail", year)
		}
	}
}

func TestCreditSecret_Validate_MissingPin(t *testing.T) {
	c := validCreditSecret()
	c.Pin = nil
	err := c.Validate()
	if err == nil {
		t.Fatal("Validate() should fail with empty PIN")
	}
	if !core.IsDomainError(err) {
		t.Errorf("expected DomainError, got %T", err)
	}
}

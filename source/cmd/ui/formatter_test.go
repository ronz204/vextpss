package ui_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"vextpss/source/cmd/ui"
	"vextpss/source/core"
	"vextpss/source/core/secrets"
)

// captureStdout redirects os.Stdout for the duration of f and returns what was written.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	old := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}

func TestPrintSecret_AccountSecret(t *testing.T) {
	payload := &secrets.AccountSecret{
		Username: "alice",
		Password: []byte("s3cr3t"),
	}

	out := captureStdout(t, func() {
		ui.PrintSecret("github", payload)
	})

	if !strings.Contains(out, "github") {
		t.Errorf("output %q should contain service name %q", out, "github")
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("output %q should contain username %q", out, "alice")
	}
	if !strings.Contains(out, "s3cr3t") {
		t.Errorf("output %q should contain password", out)
	}
}

func TestPrintSecret_UnknownType(t *testing.T) {
	// A payload whose type is not *secrets.AccountSecret falls through to the default case.
	var unknown core.SecretPayload = &unknownPayload{}

	out := captureStdout(t, func() {
		ui.PrintSecret("svc", unknown)
	})

	if !strings.Contains(out, "svc") {
		t.Errorf("output %q should contain service name", out)
	}
	if !strings.Contains(out, "unknown type") {
		t.Errorf("output %q should contain 'unknown type'", out)
	}
}

func TestPrintSecretList_Empty(t *testing.T) {
	out := captureStdout(t, func() {
		ui.PrintSecretList([]core.Secret{})
	})

	if !strings.Contains(out, "No secrets stored") {
		t.Errorf("output %q should mention no secrets stored", out)
	}
}

func TestPrintSecretList_WithSecrets(t *testing.T) {
	secrets := []core.Secret{
		{Name: "github", Type: "account", CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)},
		{Name: "gitlab", Type: "account", CreatedAt: time.Date(2024, 2, 20, 12, 0, 0, 0, time.UTC)},
	}

	out := captureStdout(t, func() {
		ui.PrintSecretList(secrets)
	})

	if !strings.Contains(out, "github") {
		t.Errorf("output %q should contain 'github'", out)
	}
	if !strings.Contains(out, "gitlab") {
		t.Errorf("output %q should contain 'gitlab'", out)
	}
	if !strings.Contains(out, "account") {
		t.Errorf("output %q should contain type 'account'", out)
	}
	if !strings.Contains(out, "NAME") {
		t.Errorf("output %q should contain column header NAME", out)
	}
}

func TestPrintSecret_CreditSecret_RequiredFields(t *testing.T) {
	payload := &secrets.CreditSecret{
		Number:          "4532123456789012",
		SecurityCode:    []byte("123"),
		ExpirationMonth: 6,
		ExpirationYear:  2028,
		Pin:             []byte("1234"),
	}

	out := captureStdout(t, func() {
		ui.PrintSecret("visa-debit", payload)
	})

	if !strings.Contains(out, "visa-debit") {
		t.Errorf("output %q should contain service name", out)
	}
	if !strings.Contains(out, "4532 1234 5678 9012") {
		t.Errorf("output %q should contain formatted card number", out)
	}
	if !strings.Contains(out, "123") {
		t.Errorf("output %q should contain security code", out)
	}
	if !strings.Contains(out, "06/2028") {
		t.Errorf("output %q should contain expiration date", out)
	}
	if !strings.Contains(out, "1234") {
		t.Errorf("output %q should contain PIN", out)
	}
}

func TestPrintSecret_CreditSecret_OptionalFieldsShown(t *testing.T) {
	payload := &secrets.CreditSecret{
		Number:          "4532123456789012",
		SecurityCode:    []byte("123"),
		ExpirationMonth: 6,
		ExpirationYear:  2028,
		Pin:             []byte("1234"),
		BankUsername:    "user@bank.com",
		BankPassword:    []byte("bankpass"),
		BankVirtualKey:  []byte("vk999"),
		BankCellphone:   "+57 300 123 4567",
		CountryCode:     "CO",
	}

	out := captureStdout(t, func() {
		ui.PrintSecret("visa-debit", payload)
	})

	if !strings.Contains(out, "user@bank.com") {
		t.Errorf("output %q should contain bank username", out)
	}
	if !strings.Contains(out, "bankpass") {
		t.Errorf("output %q should contain bank password", out)
	}
	if !strings.Contains(out, "vk999") {
		t.Errorf("output %q should contain virtual key", out)
	}
	if !strings.Contains(out, "+57 300 123 4567") {
		t.Errorf("output %q should contain cellphone", out)
	}
	if !strings.Contains(out, "CO") {
		t.Errorf("output %q should contain country code", out)
	}
}

func TestPrintSecret_CreditSecret_OptionalFieldsOmitted(t *testing.T) {
	payload := &secrets.CreditSecret{
		Number:          "4532123456789012",
		SecurityCode:    []byte("123"),
		ExpirationMonth: 6,
		ExpirationYear:  2028,
		Pin:             []byte("1234"),
	}

	out := captureStdout(t, func() {
		ui.PrintSecret("visa-debit", payload)
	})

	if strings.Contains(out, "Bank Username:") {
		t.Errorf("output %q should not contain Bank Username: when bank username is empty", out)
	}
}

// unknownPayload is a test stub for a secret type the formatter does not know.
type unknownPayload struct{}

func (u *unknownPayload) GetType() string                     { return "unknown" }
func (u *unknownPayload) Validate() error                     { return nil }
func (u *unknownPayload) MergeFrom(_ core.SecretPayload)      {}

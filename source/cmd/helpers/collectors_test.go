package helpers_test

import (
	"errors"
	"strings"
	"testing"

	"vextpss/source/cmd/helpers"
	"vextpss/source/core/secrets"
)

// --- AccountCollector ---

func TestAccountCollector_Collect_Success(t *testing.T) {
	prompter := &helpers.MockPrompter{
		LineResponse:  "alice",
		PasswordBytes: []byte("s3cr3t"),
	}
	c, _ := helpers.NewSecretCollector("account", helpers.CollectorOptions{})
	payload, err := c.Collect(prompter)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	acc, ok := payload.(*secrets.AccountSecret)
	if !ok {
		t.Fatalf("payload type = %T, want *secrets.AccountSecret", payload)
	}
	if acc.Username != "alice" {
		t.Errorf("username = %q, want %q", acc.Username, "alice")
	}
}

func TestAccountCollector_Collect_ReadLineError(t *testing.T) {
	prompter := &helpers.MockPrompter{Err: errors.New("stdin closed")}
	c, _ := helpers.NewSecretCollector("account", helpers.CollectorOptions{})
	_, err := c.Collect(prompter)
	if err == nil {
		t.Fatal("Collect() should return error when ReadLine fails")
	}
}

func TestAccountCollector_Collect_ReadPasswordError(t *testing.T) {
	prompter := &sequentialCollectorPrompter{
		lineResponses:     []string{"alice"},
		passwordResponses: []collectorPasswordResponse{{nil, errors.New("no tty")}},
	}
	c, _ := helpers.NewSecretCollector("account", helpers.CollectorOptions{})
	_, err := c.Collect(prompter)
	if err == nil {
		t.Fatal("Collect() should return error when ReadPassword fails")
	}
}

func TestAccountCollector_Collect_WithGenerate(t *testing.T) {
	prompter := &helpers.MockPrompter{LineResponse: "alice"}
	c, _ := helpers.NewSecretCollector("account", helpers.CollectorOptions{Generate: true, GenLength: 16, GenSymbols: true})
	payload, err := c.Collect(prompter)
	if err != nil {
		t.Fatalf("Collect() with generate error = %v", err)
	}
	acc := payload.(*secrets.AccountSecret)
	if len(acc.Password) == 0 {
		t.Error("generated password should not be empty")
	}
}

// --- CreditCollector ---

func newCreditPrompter(lines []string, passwords [][]byte) helpers.Prompter {
	var pwResponses []collectorPasswordResponse
	for _, p := range passwords {
		pwResponses = append(pwResponses, collectorPasswordResponse{p, nil})
	}
	return &sequentialCollectorPrompter{
		lineResponses:     lines,
		passwordResponses: pwResponses,
	}
}

func TestCreditCollector_Collect_Success(t *testing.T) {
	lines := []string{
		"4532123456789012", // card number
		"6",               // expiration month
		"2028",            // expiration year
		"user@bank.com",   // bank username
		"+57 300 123",     // cellphone
		"CO",              // country code
	}
	passwords := [][]byte{
		[]byte("123"),     // security code
		[]byte("1234"),    // PIN
		[]byte("bankp4s"), // bank password (prompted because bank username is set)
		[]byte("vk999"),   // virtual key
	}

	c, _ := helpers.NewSecretCollector("credit", helpers.CollectorOptions{})
	payload, err := c.Collect(newCreditPrompter(lines, passwords))
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	credit, ok := payload.(*secrets.CreditSecret)
	if !ok {
		t.Fatalf("payload type = %T, want *secrets.CreditSecret", payload)
	}
	if credit.Number != "4532123456789012" {
		t.Errorf("Number = %q, want %q", credit.Number, "4532123456789012")
	}
	if credit.ExpirationMonth != 6 {
		t.Errorf("ExpirationMonth = %d, want 6", credit.ExpirationMonth)
	}
	if credit.ExpirationYear != 2028 {
		t.Errorf("ExpirationYear = %d, want 2028", credit.ExpirationYear)
	}
	if string(credit.SecurityCode) != "123" {
		t.Errorf("SecurityCode = %q, want %q", credit.SecurityCode, "123")
	}
}

func TestCreditCollector_Collect_OptionalFieldsSkipped(t *testing.T) {
	lines := []string{
		"4532123456789012",
		"12",   // month
		"2030", // year
		"",     // bank username (skip) — no bank password prompt
		"",     // cellphone
		"",     // country
	}
	passwords := [][]byte{
		[]byte("456"),  // security code
		[]byte("0000"), // PIN
		[]byte(""),     // virtual key (empty = skipped)
	}

	c, _ := helpers.NewSecretCollector("credit", helpers.CollectorOptions{})
	payload, err := c.Collect(newCreditPrompter(lines, passwords))
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	credit := payload.(*secrets.CreditSecret)
	if credit.BankUsername != "" {
		t.Errorf("BankUsername should be empty, got %q", credit.BankUsername)
	}
	if len(credit.BankPassword) != 0 {
		t.Errorf("BankPassword should be empty, got %q", credit.BankPassword)
	}
}

func TestCreditCollector_Collect_InvalidMonth(t *testing.T) {
	lines := []string{
		"4532123456789012",
		"13",   // invalid month
		"2028",
		"", "", "",
	}
	passwords := [][]byte{[]byte("123"), []byte("1234"), []byte("")}

	c, _ := helpers.NewSecretCollector("credit", helpers.CollectorOptions{})
	_, err := c.Collect(newCreditPrompter(lines, passwords))
	if err == nil {
		t.Fatal("Collect() should return error for invalid month")
	}
	if !strings.Contains(err.Error(), "month") {
		t.Errorf("error %q should mention 'month'", err.Error())
	}
}

func TestCreditCollector_Collect_CardNumberSpacesStripped(t *testing.T) {
	lines := []string{
		"4532 1234 5678 9012", // number with spaces
		"6", "2028", "", "", "",
	}
	passwords := [][]byte{[]byte("123"), []byte("1234"), []byte("")}

	c, _ := helpers.NewSecretCollector("credit", helpers.CollectorOptions{})
	payload, err := c.Collect(newCreditPrompter(lines, passwords))
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	credit := payload.(*secrets.CreditSecret)
	if credit.Number != "4532123456789012" {
		t.Errorf("Number spaces not stripped, got %q", credit.Number)
	}
}

// --- Factory ---

func TestNewSecretCollector_UnknownType(t *testing.T) {
	_, err := helpers.NewSecretCollector("note", helpers.CollectorOptions{})
	if err == nil {
		t.Fatal("NewSecretCollector() should return error for unknown type")
	}
}

// sequentialCollectorPrompter returns pre-set responses in order.
type collectorPasswordResponse struct {
	bytes []byte
	err   error
}

type sequentialCollectorPrompter struct {
	lineResponses     []string
	passwordResponses []collectorPasswordResponse
	lineIdx           int
	passwordIdx       int
}

func (s *sequentialCollectorPrompter) ReadLine(_ string) (string, error) {
	if s.lineIdx >= len(s.lineResponses) {
		return "", nil
	}
	resp := s.lineResponses[s.lineIdx]
	s.lineIdx++
	return resp, nil
}

func (s *sequentialCollectorPrompter) ReadPassword(_ string) ([]byte, error) {
	if s.passwordIdx >= len(s.passwordResponses) {
		return []byte(""), nil
	}
	resp := s.passwordResponses[s.passwordIdx]
	s.passwordIdx++
	return resp.bytes, resp.err
}

func (s *sequentialCollectorPrompter) Confirm(_ string) (bool, error) { return false, nil }

func (s *sequentialCollectorPrompter) Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

package shared_test

import (
	"strings"
	"testing"
	"unicode"

	"vextpss/source/pkg/shared"
)

func TestGeneratePassword_Length(t *testing.T) {
	for _, length := range []int{1, 8, 16, 32, 64} {
		pw, err := shared.GeneratePassword(length, true)
		if err != nil {
			t.Fatalf("GeneratePassword(%d) error = %v", length, err)
		}
		if len(pw) != length {
			t.Errorf("GeneratePassword(%d) len = %d, want %d", length, len(pw), length)
		}
	}
}

func TestGeneratePassword_NoSymbols_ContainsOnlyAlphanumeric(t *testing.T) {
	pw, err := shared.GeneratePassword(200, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, ch := range pw {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			t.Errorf("GeneratePassword with no-symbols contains non-alphanumeric char %q", ch)
		}
	}
}

func TestGeneratePassword_WithSymbols_ContainsSymbol(t *testing.T) {
	// With 200 chars and a 26+26+10+27 = 89-char pool, P(no symbol in 200 chars) ≈ (62/89)^200 ≈ 0
	pw, err := shared.GeneratePassword(200, true)
	if err != nil {
		t.Fatal(err)
	}
	const symbols = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	hasSymbol := strings.ContainsAny(pw, symbols)
	if !hasSymbol {
		t.Error("GeneratePassword with symbols should contain at least one symbol in 200 chars")
	}
}

func TestGeneratePassword_Uniqueness(t *testing.T) {
	a, _ := shared.GeneratePassword(20, true)
	b, _ := shared.GeneratePassword(20, true)
	if a == b {
		t.Error("two separate GeneratePassword calls should not return identical passwords")
	}
}

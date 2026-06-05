package shared_test

import (
	"bytes"
	"testing"

	"vextpss/source/pkg/shared"
)

func TestZero_ClearsAllBytes(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5}
	shared.Zero(b)
	for i, v := range b {
		if v != 0 {
			t.Errorf("byte at index %d = %d, want 0", i, v)
		}
	}
}

func TestZero_EmptySlice(t *testing.T) {
	shared.Zero([]byte{}) // must not panic
}

func TestZero_NilSlice(t *testing.T) {
	shared.Zero(nil) // must not panic
}

func TestGenerateSalt_Returns16Bytes(t *testing.T) {
	salt, err := shared.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}
	if len(salt) != 16 {
		t.Errorf("len(salt) = %d, want 16", len(salt))
	}
}

func TestGenerateSalt_IsRandom(t *testing.T) {
	a, err := shared.GenerateSalt()
	if err != nil {
		t.Fatalf("first GenerateSalt() error = %v", err)
	}
	b, err := shared.GenerateSalt()
	if err != nil {
		t.Fatalf("second GenerateSalt() error = %v", err)
	}
	if bytes.Equal(a, b) {
		t.Error("two consecutive salts should not be equal")
	}
}

func TestNow_ReturnsUTC(t *testing.T) {
	now := shared.Now()
	if now.Location() != now.UTC().Location() {
		t.Errorf("Now() location = %v, want UTC", now.Location())
	}
}

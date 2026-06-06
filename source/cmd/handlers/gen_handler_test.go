package handlers_test

import (
	"strings"
	"testing"

	"vextpss/source/cmd/handlers"
)

func TestGenHandler_Handle_DefaultLength(t *testing.T) {
	h := handlers.NewGenHandler()
	out := captureStdout(t, func() {
		if err := h.Handle(20, true); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})
	pw := strings.TrimSpace(out)
	if len(pw) != 20 {
		t.Errorf("generated password length = %d, want 20", len(pw))
	}
}

func TestGenHandler_Handle_CustomLength(t *testing.T) {
	h := handlers.NewGenHandler()
	out := captureStdout(t, func() {
		h.Handle(32, true) //nolint
	})
	pw := strings.TrimSpace(out)
	if len(pw) != 32 {
		t.Errorf("generated password length = %d, want 32", len(pw))
	}
}

func TestGenHandler_Handle_ZeroLength(t *testing.T) {
	h := handlers.NewGenHandler()
	out := captureStdout(t, func() {
		if err := h.Handle(0, true); err != nil {
			t.Errorf("Handle() should not return an error, got %v", err)
		}
	})
	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X] for length 0", out)
	}
}

func TestGenHandler_Handle_NoSymbols(t *testing.T) {
	h := handlers.NewGenHandler()
	out := captureStdout(t, func() {
		h.Handle(100, false) //nolint
	})
	pw := strings.TrimSpace(out)
	const symbols = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	if strings.ContainsAny(pw, symbols) {
		t.Errorf("no-symbols password %q contains symbol characters", pw)
	}
}

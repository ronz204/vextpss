package core_test

import (
	"errors"
	"fmt"
	"testing"

	"vextpss/source/core"
)

func TestDomainError_Error(t *testing.T) {
	err := core.NewDomainError("something went wrong")
	if got := err.Error(); got != "something went wrong" {
		t.Errorf("Error() = %q, want %q", got, "something went wrong")
	}
}

func TestDomainError_IsError(t *testing.T) {
	var err error = core.NewDomainError("msg")
	if err == nil {
		t.Error("DomainError should satisfy the error interface")
	}
}

func TestCanonicalErrors_HaveExpectedMessages(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{core.ErrSecretNotFound, "secret not found"},
		{core.ErrAlreadyExists, "secret already exists"},
		{core.ErrInvalidCredential, "invalid credential format"},
		{core.ErrDecryptionFailed, "master password incorrect or data corrupted"},
		{core.ErrInvalidInput, "invalid input"},
	}
	for _, tt := range tests {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("error message = %q, want %q", got, tt.want)
		}
	}
}

func TestIsDomainError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"direct domain error", core.NewDomainError("msg"), true},
		{"canonical not found", core.ErrSecretNotFound, true},
		{"wrapped domain error", fmt.Errorf("wrap: %w", core.NewDomainError("msg")), true},
		{"stdlib error", errors.New("std"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := core.IsDomainError(tt.err); got != tt.want {
				t.Errorf("IsDomainError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrSecretNotFound", core.ErrSecretNotFound, true},
		{"wrapped not found", fmt.Errorf("layer: %w", core.ErrSecretNotFound), true},
		{"ErrAlreadyExists", core.ErrAlreadyExists, false},
		{"stdlib error", errors.New("not found"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := core.IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrAlreadyExists", core.ErrAlreadyExists, true},
		{"wrapped already exists", fmt.Errorf("layer: %w", core.ErrAlreadyExists), true},
		{"ErrSecretNotFound", core.ErrSecretNotFound, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := core.IsAlreadyExists(tt.err); got != tt.want {
				t.Errorf("IsAlreadyExists(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

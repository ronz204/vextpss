package domain_test

import (
	"testing"
	"vextpss/source/pkg/domain"
)

func TestAccountSecretValidate(t *testing.T) {
	tests := []struct {
		name    string
		secret  domain.AccountSecret
		wantErr bool
	}{
		{
			name:    "valid",
			secret:  domain.AccountSecret{Username: "alice", Password: []byte("s3cr3t")},
			wantErr: false,
		},
		{
			name:    "empty username",
			secret:  domain.AccountSecret{Username: "", Password: []byte("s3cr3t")},
			wantErr: true,
		},
		{
			name:    "empty password",
			secret:  domain.AccountSecret{Username: "alice", Password: []byte{}},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.secret.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestDomainErrors(t *testing.T) {
	if !domain.IsDomainError(domain.ErrSecretNotFound) {
		t.Error("ErrSecretNotFound should be a DomainError")
	}
	if !domain.IsDomainError(domain.ErrAlreadyExists) {
		t.Error("ErrAlreadyExists should be a DomainError")
	}
	if !domain.IsNotFound(domain.ErrSecretNotFound) {
		t.Error("IsNotFound should return true for ErrSecretNotFound")
	}
	if !domain.IsAlreadyExists(domain.ErrAlreadyExists) {
		t.Error("IsAlreadyExists should return true for ErrAlreadyExists")
	}
}

func TestAccountSecretGetType(t *testing.T) {
	s := &domain.AccountSecret{}
	if s.GetType() != "account" {
		t.Errorf("expected type 'account', got %q", s.GetType())
	}
}

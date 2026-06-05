package secrets_test

import (
	"strings"
	"testing"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
)

func TestAccountSecret_GetType(t *testing.T) {
	s := &secrets.AccountSecret{Username: "user", Password: []byte("pass")}
	if got := s.GetType(); got != "account" {
		t.Errorf("GetType() = %q, want %q", got, "account")
	}
}

func TestAccountSecret_Validate(t *testing.T) {
	tests := []struct {
		name      string
		secret    secrets.AccountSecret
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid account",
			secret:  secrets.AccountSecret{Username: "alice", Password: []byte("s3cr3t")},
			wantErr: false,
		},
		{
			name:      "empty username",
			secret:    secrets.AccountSecret{Username: "", Password: []byte("pass")},
			wantErr:   true,
			errSubstr: "username is required",
		},
		{
			name:      "empty password",
			secret:    secrets.AccountSecret{Username: "alice", Password: []byte{}},
			wantErr:   true,
			errSubstr: "password is required",
		},
		{
			name:      "nil password",
			secret:    secrets.AccountSecret{Username: "alice", Password: nil},
			wantErr:   true,
			errSubstr: "password is required",
		},
		{
			name: "username at max length",
			secret: secrets.AccountSecret{
				Username: strings.Repeat("a", 255),
				Password: []byte("pass"),
			},
			wantErr: false,
		},
		{
			name: "username exceeds max length",
			secret: secrets.AccountSecret{
				Username: strings.Repeat("a", 256),
				Password: []byte("pass"),
			},
			wantErr:   true,
			errSubstr: "username exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.secret.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if !core.IsDomainError(err) {
					t.Errorf("Validate() error should be a DomainError, got %T", err)
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("Validate() error %q should contain %q", err.Error(), tt.errSubstr)
				}
			}
		})
	}
}

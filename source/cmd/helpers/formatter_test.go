package helpers_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"vextpss/source/cmd/helpers"
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
		helpers.PrintSecret("github", payload)
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
		helpers.PrintSecret("svc", unknown)
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
		helpers.PrintSecretList([]core.Secret{})
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
		helpers.PrintSecretList(secrets)
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

// unknownPayload is a test stub for a secret type the formatter does not know.
type unknownPayload struct{}

func (u *unknownPayload) GetType() string  { return "unknown" }
func (u *unknownPayload) Validate() error  { return nil }

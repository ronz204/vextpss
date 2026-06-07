package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vextpss/source/cmd/handlers"
	"vextpss/source/app"
	"vextpss/source/testutil"
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

func newInitUC(t *testing.T, initErr error) *app.InitStorageUC {
	t.Helper()
	var initFn func(ctx context.Context) error
	if initErr != nil {
		initFn = func(_ context.Context) error { return initErr }
	}

	dbPath := ""
	if initErr == nil {
		dir := t.TempDir()
		dbPath = filepath.Join(dir, "vext.db")
		f, err := os.Create(dbPath)
		if err != nil {
			t.Fatalf("could not create temp db: %v", err)
		}
		f.Close()
	}

	return app.NewInitStorageUC(&testutil.MockInitialiser{
		InitFn:    initFn,
		DBPathVal: dbPath,
	})
}

func TestInitHandler_Handle_Success(t *testing.T) {
	uc := newInitUC(t, nil)
	h := handlers.NewInitHandler(uc)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background()); err != nil {
			t.Errorf("Handle() error = %v", err)
		}
	})

	if !strings.Contains(out, "initialized") {
		t.Errorf("output %q should contain 'initialized'", out)
	}
}

func TestInitHandler_Handle_Error(t *testing.T) {
	uc := newInitUC(t, errors.New("permission denied"))
	h := handlers.NewInitHandler(uc)

	out := captureStdout(t, func() {
		if err := h.Handle(context.Background()); err != nil {
			t.Errorf("Handle() should not return error, got %v", err)
		}
	})

	if !strings.Contains(out, "[X]") {
		t.Errorf("output %q should contain error marker [X]", out)
	}
}

package apps_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"vextpss/source/pkg/apps"
	"vextpss/source/tests/mocks"
)

func TestInitStorageUC_Execute_Success(t *testing.T) {
	dir := t.TempDir()
	dbFile := filepath.Join(dir, "vext.db")

	// Create the file so Chmod has something to act on.
	f, err := os.Create(dbFile)
	if err != nil {
		t.Fatalf("could not create temp db file: %v", err)
	}
	f.Close()

	init := &mocks.MockInitialiser{DBPathVal: dbFile}
	uc := apps.NewInitStorageUC(init)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestInitStorageUC_Execute_InitError(t *testing.T) {
	initErr := errors.New("could not create directory")
	init := &mocks.MockInitialiser{
		InitFn: func(_ context.Context) error { return initErr },
	}
	uc := apps.NewInitStorageUC(init)

	err := uc.Execute(context.Background())
	if err == nil {
		t.Fatal("Execute() with init error should return an error")
	}
}

func TestInitStorageUC_Execute_ChmodError(t *testing.T) {
	// Init succeeds but DBPath points to a non-existent file → Chmod fails.
	init := &mocks.MockInitialiser{DBPathVal: "/nonexistent/path/vext.db"}
	uc := apps.NewInitStorageUC(init)

	err := uc.Execute(context.Background())
	if err == nil {
		t.Fatal("Execute() with missing db file should return a chmod error")
	}
}

package repos_test

import (
	"context"
	"testing"
	"time"

	"vextpss/source/core"
	"vextpss/source/dal"
	"vextpss/source/dal/repos"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := db.AutoMigrate(&dal.SecretRecord{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func makeSecret(name, typ string) *core.Secret {
	return &core.Secret{
		Name:      name,
		Type:      typ,
		Salt:      []byte("salt-16-bytes!!x"),
		Nonce:     []byte("nonce-12-byt"),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func TestSQLiteRepository_Save_Success(t *testing.T) {
	repo := repos.NewSQLiteRepository(newTestDB(t))
	secret := makeSecret("github", "account")
	encrypted := []byte("encrypted-payload")

	if err := repo.Save(context.Background(), secret, encrypted); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
}

func TestSQLiteRepository_Save_DuplicateName(t *testing.T) {
	repo := repos.NewSQLiteRepository(newTestDB(t))
	secret := makeSecret("github", "account")

	if err := repo.Save(context.Background(), secret, []byte("enc1")); err != nil {
		t.Fatalf("first Save() error = %v", err)
	}

	err := repo.Save(context.Background(), makeSecret("github", "account"), []byte("enc2"))
	if !core.IsAlreadyExists(err) {
		t.Errorf("second Save() error = %v, want ErrAlreadyExists", err)
	}
}

func TestSQLiteRepository_GetByName_Success(t *testing.T) {
	repo := repos.NewSQLiteRepository(newTestDB(t))
	secret := makeSecret("github", "account")
	encrypted := []byte("encrypted-data")

	if err := repo.Save(context.Background(), secret, encrypted); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, gotEnc, err := repo.GetByName(context.Background(), "github")
	if err != nil {
		t.Fatalf("GetByName() error = %v", err)
	}
	if got.Name != "github" {
		t.Errorf("Name = %q, want %q", got.Name, "github")
	}
	if got.Type != "account" {
		t.Errorf("Type = %q, want %q", got.Type, "account")
	}
	if string(gotEnc) != string(encrypted) {
		t.Errorf("encrypted = %q, want %q", gotEnc, encrypted)
	}
}

func TestSQLiteRepository_GetByName_NotFound(t *testing.T) {
	repo := repos.NewSQLiteRepository(newTestDB(t))

	_, _, err := repo.GetByName(context.Background(), "missing")
	if !core.IsNotFound(err) {
		t.Errorf("GetByName() error = %v, want ErrSecretNotFound", err)
	}
}

func TestSQLiteRepository_GetByName_PreservesFields(t *testing.T) {
	db := newTestDB(t)
	repo := repos.NewSQLiteRepository(db)
	secret := makeSecret("gitlab", "account")
	secret.Salt = []byte("specific-salt123")
	secret.Nonce = []byte("specific-nonce!")

	if err := repo.Save(context.Background(), secret, []byte("enc")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, _, err := repo.GetByName(context.Background(), "gitlab")
	if err != nil {
		t.Fatalf("GetByName() error = %v", err)
	}
	if string(got.Salt) != string(secret.Salt) {
		t.Errorf("Salt = %q, want %q", got.Salt, secret.Salt)
	}
	if string(got.Nonce) != string(secret.Nonce) {
		t.Errorf("Nonce = %q, want %q", got.Nonce, secret.Nonce)
	}
}

func TestSQLiteRepository_ListAll_ReturnsOrderedByName(t *testing.T) {
	repo := repos.NewSQLiteRepository(newTestDB(t))
	enc := []byte("enc")

	for _, name := range []string{"zabbix", "ansible", "mattermost"} {
		if err := repo.Save(context.Background(), makeSecret(name, "account"), enc); err != nil {
			t.Fatalf("Save(%q) error = %v", name, err)
		}
	}

	all, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("len(all) = %d, want 3", len(all))
	}

	expected := []string{"ansible", "mattermost", "zabbix"}
	for i, s := range all {
		if s.Name != expected[i] {
			t.Errorf("all[%d].Name = %q, want %q", i, s.Name, expected[i])
		}
	}
}

func TestSQLiteRepository_ListAll_Empty(t *testing.T) {
	repo := repos.NewSQLiteRepository(newTestDB(t))

	all, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty list, got %d items", len(all))
	}
}

func TestSQLiteRepository_ListAll_DoesNotReturnEncryptedPayload(t *testing.T) {
	repo := repos.NewSQLiteRepository(newTestDB(t))
	secret := makeSecret("svc", "account")

	if err := repo.Save(context.Background(), secret, []byte("sensitive-encrypted-data")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	all, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least one result")
	}
	// ListAll returns core.Secret (no Encrypted field) — just verify no panic and correct type.
	if all[0].Name != "svc" {
		t.Errorf("Name = %q, want %q", all[0].Name, "svc")
	}
}

func TestSQLiteRepository_Delete_Success(t *testing.T) {
	repo := repos.NewSQLiteRepository(newTestDB(t))
	if err := repo.Save(context.Background(), makeSecret("github", "account"), []byte("enc")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := repo.Delete(context.Background(), "github"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, _, err := repo.GetByName(context.Background(), "github")
	if !core.IsNotFound(err) {
		t.Errorf("after Delete, GetByName error = %v, want ErrSecretNotFound", err)
	}
}

func TestSQLiteRepository_Delete_NotFound(t *testing.T) {
	repo := repos.NewSQLiteRepository(newTestDB(t))

	err := repo.Delete(context.Background(), "nonexistent")
	if !core.IsNotFound(err) {
		t.Errorf("Delete() error = %v, want ErrSecretNotFound", err)
	}
}

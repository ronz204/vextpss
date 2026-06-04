package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"vextpss/source/pkg/domain"
	"vextpss/source/pkg/models"

	"gorm.io/gorm"
)

// SQLiteRepository implements domain.SecretRepository using GORM + SQLite.
type SQLiteRepository struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// Save persists a secret with its pre-encrypted payload.
func (r *SQLiteRepository) Save(
	ctx context.Context,
	secret *domain.Secret,
	encryptedPayload []byte,
) error {
	record := &models.SecretRecord{
		Name:             secret.Name,
		Type:             secret.Type,
		Salt:             secret.Salt,
		Nonce:            secret.Nonce,
		EncryptedPayload: encryptedPayload,
	}

	result := r.db.WithContext(ctx).Create(record)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("insert failed: %w", result.Error)
	}
	return nil
}

// GetByName retrieves a secret by name, returning its encrypted payload separately.
func (r *SQLiteRepository) GetByName(
	ctx context.Context,
	name string,
) (*domain.Secret, []byte, error) {
	var record models.SecretRecord
	result := r.db.WithContext(ctx).Where("name = ?", name).First(&record)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil, domain.ErrSecretNotFound
	}
	if result.Error != nil {
		return nil, nil, fmt.Errorf("query failed: %w", result.Error)
	}

	secret := &domain.Secret{
		ID:        domain.SecretID(record.ID),
		Name:      record.Name,
		Type:      record.Type,
		Salt:      record.Salt,
		Nonce:     record.Nonce,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}

	return secret, record.EncryptedPayload, nil
}

// ListAll returns metadata for every secret without touching the encrypted payload.
func (r *SQLiteRepository) ListAll(ctx context.Context) ([]domain.Secret, error) {
	var records []models.SecretRecord
	result := r.db.WithContext(ctx).
		Select("id, name, type, created_at, updated_at").
		Order("name asc").
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("list query failed: %w", result.Error)
	}

	secrets := make([]domain.Secret, len(records))
	for i, rec := range records {
		secrets[i] = domain.Secret{
			ID:        domain.SecretID(rec.ID),
			Name:      rec.Name,
			Type:      rec.Type,
			CreatedAt: rec.CreatedAt,
			UpdatedAt: rec.UpdatedAt,
		}
	}
	return secrets, nil
}

// Delete removes a secret by name. Returns ErrSecretNotFound when absent.
func (r *SQLiteRepository) Delete(ctx context.Context, name string) error {
	result := r.db.WithContext(ctx).Where("name = ?", name).Delete(&models.SecretRecord{})
	if result.Error != nil {
		return fmt.Errorf("delete failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrSecretNotFound
	}
	return nil
}

// isUniqueConstraintError detects SQLite UNIQUE constraint violations.
func isUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// --- StorageInitialiser implementation ---

// Initialiser wraps the storage setup needed by InitStorageUC.
type Initialiser struct {
	dbPath string
	appDir string
}

func NewInitialiser(dbPath, appDir string) *Initialiser {
	return &Initialiser{dbPath: dbPath, appDir: appDir}
}

func (i *Initialiser) DBPath() string { return i.dbPath }

// Init creates the app directory, opens/creates the database, and applies migrations.
func (i *Initialiser) Init(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(i.dbPath), 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	db, err := Open(i.dbPath)
	if err != nil {
		return err
	}
	defer Close(db)

	return Migrate(db)
}

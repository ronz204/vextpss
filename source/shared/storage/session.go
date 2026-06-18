package storage

import "vextpss/source/core"

// WithRepo opens the database, creates a repository, executes fn, then closes the database.
// The connection is always closed regardless of whether fn returns an error.
func WithRepo(dbPath string, fn func(core.SecretRepository) error) error {
	db, err := Open(dbPath)
	if err != nil {
		return err
	}
	defer Close(db)
	return fn(NewSecretRepository(db))
}

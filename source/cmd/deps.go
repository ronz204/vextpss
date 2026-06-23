package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/collectors"
	"vextpss/source/cmd/formatters"
	"vextpss/source/secrets"
	"vextpss/source/shared/cryptors"
	"vextpss/source/shared/storage"
)

// Deps holds the resolved infrastructure for a single command execution.
// All fields are interfaces — no concrete infra types leak into the command layer.
type Deps struct {
	Repo    secrets.Repository
	Enc     secrets.Encryptor
	Collect *collectors.Collector
}

// Handler is the signature all command implementations must satisfy.
type Handler func(ctx context.Context, d Deps, args []string) error

// Middleware decorates a Handler with cross-cutting behavior.
type Middleware func(Handler) Handler

// chain composes middleware left-to-right (first = outermost).
func chain(mw ...Middleware) Middleware {
	return func(h Handler) Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}
		return h
	}
}

// withDeps resolves all command dependencies, opens the database,
// applies any middleware, executes the handler, and closes the database.
// It is the composition root for every command that requires infrastructure.
func withDeps(h Handler, mw ...Middleware) func(*cobra.Command, []string) error {
	h = chain(mw...)(h)
	return func(_ *cobra.Command, args []string) error {
		dbPath, err := storage.DBPath()
		if err != nil {
			formatters.Error(err.Error())
			return err

		}
		db, err := storage.Open(dbPath)
		if err != nil {
			formatters.Error("storage not initialized — run `vext init` first")
			return err
		}
		defer storage.Close(db)

		return h(context.Background(), Deps{
			Repo:    storage.NewSecretRepository(db),
			Enc:     cryptors.NewAESGCMEncryptor(cryptors.DefaultConfig()),
			Collect: collectors.NewCollector(collectors.NewPrompter()),
		}, args)
	}
}

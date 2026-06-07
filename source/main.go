package main

import (
	"fmt"
	"os"

	"vextpss/source/app"
	"vextpss/source/cmd"
	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/ui"
	"vextpss/source/config"
	"vextpss/source/crypto"
	"vextpss/source/storage"
	"vextpss/source/storage/sqlite"
)

func main() {
	// 1. Load configuration (paths, env).
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 2. Open the database (single connection for the lifetime of the process).
	db, err := storage.Open(cfg.DBPath)
	if err != nil {
		// Database may not exist yet — that is fine; `vext init` creates it.
		// Only hard-fail if the file exists but cannot be opened.
		if _, statErr := os.Stat(cfg.DBPath); statErr == nil {
			fmt.Fprintf(os.Stderr, "Error: could not open database: %v\n", err)
			os.Exit(1)
		}
		db = nil
	}
	if db != nil {
		defer storage.Close(db)
	}

	// 3. Wire up infrastructure implementations.
	encryptor := crypto.NewAESGCMEncryptor()
	initialiser := storage.NewInitialiser(cfg.DBPath)

	// 4. Build use cases (repo may be nil when the DB has not been initialised yet).
	initUC := app.NewInitStorageUC(initialiser)

	var (
		storeUC    *app.StoreSecretUC
		retrieveUC *app.RetrieveSecretUC
		listUC     *app.ListSecretsUC
		deleteUC   *app.DeleteSecretUC
		updateUC   *app.UpdateSecretUC
		exportUC   *app.ExportSecretsUC
		importUC   *app.ImportSecretsUC
	)
	if db != nil {
		repo := sqlite.NewSQLiteRepository(db)
		storeUC = app.NewStoreSecretUC(repo, encryptor)
		retrieveUC = app.NewRetrieveSecretUC(repo, encryptor)
		listUC = app.NewListSecretsUC(repo)
		deleteUC = app.NewDeleteSecretUC(repo)
		updateUC = app.NewUpdateSecretUC(repo, encryptor)
		exportUC = app.NewExportSecretsUC(repo, encryptor)
		importUC = app.NewImportSecretsUC(repo, encryptor)
	}

	// 5. Build CLI handlers.
	prompter := &ui.CLIPrompter{}

	initHandler := handlers.NewInitHandler(initUC)
	addHandler := handlers.NewAddHandler(storeUC, prompter)
	getHandler := handlers.NewGetHandler(retrieveUC, prompter)
	listHandler := handlers.NewListHandler(listUC)
	rmHandler := handlers.NewRmHandler(deleteUC, prompter)
	genHandler := handlers.NewGenHandler()
	updateHandler := handlers.NewUpdateHandler(retrieveUC, updateUC, prompter)
	exportHandler := handlers.NewExportHandler(exportUC, prompter)
	importHandler := handlers.NewImportHandler(importUC, prompter)

	// 6. Assemble command tree.
	rootCmd := cmd.NewRootCmd()
	rootCmd.AddCommand(initHandler.CobraCommand())
	rootCmd.AddCommand(addHandler.CobraCommand())
	rootCmd.AddCommand(getHandler.CobraCommand())
	rootCmd.AddCommand(listHandler.CobraCommand())
	rootCmd.AddCommand(rmHandler.CobraCommand())
	rootCmd.AddCommand(genHandler.CobraCommand())
	rootCmd.AddCommand(updateHandler.CobraCommand())
	rootCmd.AddCommand(exportHandler.CobraCommand())
	rootCmd.AddCommand(importHandler.CobraCommand())

	// 7. Execute.
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

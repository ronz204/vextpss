package main

import (
	"fmt"
	"os"

	"vextpss/source/cmd"
	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/ui"
	"vextpss/source/pkg/application"
	infracrypto "vextpss/source/pkg/infrastructure/crypto"
	"vextpss/source/pkg/infrastructure/config"
	"vextpss/source/pkg/infrastructure/storage"
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
	encryptor := infracrypto.NewAESGCMEncryptor()
	initialiser := storage.NewInitialiser(cfg.DBPath, cfg.AppDir)

	// 4. Build use cases (repo may be nil when the DB has not been initialised yet).
	initUC := application.NewInitStorageUC(initialiser)

	var (
		storeUC    *application.StoreSecretUC
		retrieveUC *application.RetrieveSecretUC
		listUC     *application.ListSecretsUC
		deleteUC   *application.DeleteSecretUC
	)
	if db != nil {
		repo := storage.NewSQLiteRepository(db)
		storeUC = application.NewStoreSecretUC(repo, encryptor)
		retrieveUC = application.NewRetrieveSecretUC(repo, encryptor)
		listUC = application.NewListSecretsUC(repo)
		deleteUC = application.NewDeleteSecretUC(repo)
	}

	// 5. Build CLI handlers.
	prompter := &ui.CLIPrompter{}

	initHandler := handlers.NewInitHandler(initUC)
	addHandler := handlers.NewAddHandler(storeUC, prompter)
	getHandler := handlers.NewGetHandler(retrieveUC, prompter)
	listHandler := handlers.NewListHandler(listUC)
	rmHandler := handlers.NewRmHandler(deleteUC, prompter)

	// 6. Assemble command tree.
	rootCmd := cmd.NewRootCmd()
	rootCmd.AddCommand(initHandler.CobraCommand())
	rootCmd.AddCommand(addHandler.CobraCommand())
	rootCmd.AddCommand(getHandler.CobraCommand())
	rootCmd.AddCommand(listHandler.CobraCommand())
	rootCmd.AddCommand(rmHandler.CobraCommand())

	// 7. Execute.
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

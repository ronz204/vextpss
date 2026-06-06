package main

import (
	"fmt"
	"os"

	"vextpss/source/cmd"
	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/helpers"
	"vextpss/source/dal"
	"vextpss/source/dal/repos"
	"vextpss/source/pkg/apps"
	"vextpss/source/pkg/configs"
	"vextpss/source/pkg/tokens"
)

func main() {
	// 1. Load configuration (paths, env).
	cfg, err := configs.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 2. Open the database (single connection for the lifetime of the process).
	db, err := dal.Open(cfg.DBPath)
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
		defer dal.Close(db)
	}

	// 3. Wire up infrastructure implementations.
	encryptor := tokens.NewAESGCMEncryptor()
	initialiser := dal.NewInitialiser(cfg.DBPath, cfg.AppDir)

	// 4. Build use cases (repo may be nil when the DB has not been initialised yet).
	initUC := apps.NewInitStorageUC(initialiser)

	var (
		storeUC    *apps.StoreSecretUC
		retrieveUC *apps.RetrieveSecretUC
		listUC     *apps.ListSecretsUC
		deleteUC   *apps.DeleteSecretUC
		updateUC   *apps.UpdateSecretUC
		exportUC   *apps.ExportSecretsUC
		importUC   *apps.ImportSecretsUC
	)
	if db != nil {
		repo := repos.NewSQLiteRepository(db)
		storeUC = apps.NewStoreSecretUC(repo, encryptor)
		retrieveUC = apps.NewRetrieveSecretUC(repo, encryptor)
		listUC = apps.NewListSecretsUC(repo)
		deleteUC = apps.NewDeleteSecretUC(repo)
		updateUC = apps.NewUpdateSecretUC(repo, encryptor)
		exportUC = apps.NewExportSecretsUC(repo, encryptor)
		importUC = apps.NewImportSecretsUC(repo, encryptor)
	}

	// 5. Build CLI handlers.
	prompter := &helpers.CLIPrompter{}

	initHandler := handlers.NewInitHandler(initUC)
	addHandler := handlers.NewAddHandler(storeUC, prompter)
	getHandler := handlers.NewGetHandler(retrieveUC, prompter)
	listHandler := handlers.NewListHandler(listUC)
	rmHandler := handlers.NewRmHandler(deleteUC, prompter)
	genHandler := handlers.NewGenHandler()
	updateHandler := handlers.NewUpdateHandler(updateUC, prompter)
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

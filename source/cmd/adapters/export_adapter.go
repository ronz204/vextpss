package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
	"vextpss/source/shared/memory"
	"vextpss/source/shared/storage"
)

func ExportCmd(d Deps) *cobra.Command {
	var outPath string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all secrets to an encrypted file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(outPath, d)
		},
	}

	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output file path (default: vext-export-YYYYMMDD-HHMMSS.vxt)")
	return cmd
}

func runExport(outPath string, d Deps) error {
	if outPath == "" {
		outPath = fmt.Sprintf("vext-export-%s.vxt", time.Now().Format("20060102-150405"))
	}

	masterPassword, err := d.Collector.Master()
	defer memory.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	err = storage.WithRepo(d.DBPath, func(repo *storage.SecretRepository) error {
		return funcs.NewExportSecretsFunc(repo, d.Encryptor).Run(context.Background(), funcs.ExportSecretsDto{
			FilePath:       outPath,
			MasterPassword: masterPassword,
		})
	})
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	formatters.Success(fmt.Sprintf("Exported to %s", outPath))
	return nil
}

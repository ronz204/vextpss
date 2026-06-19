package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
	"vextpss/source/shared"
	"vextpss/source/shared/passgen"
	"vextpss/source/shared/storage"
)

// ================================
// ExportCmd returns the cobra command for "vext export".
// ================================
func ExportCmd(deps shared.AppDeps) *cobra.Command {
	var outPath string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all secrets to an encrypted file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(outPath, deps)
		},
	}

	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output file path (default: vext-export-YYYYMMDD-HHMMSS.vxt)")
	return cmd
}

func runExport(outPath string, deps shared.AppDeps) error {
	if outPath == "" {
		outPath = fmt.Sprintf("vext-export-%s.vxt", time.Now().Format("20060102-150405"))
	}

	masterPassword, err := deps.Collector.Master()
	defer passgen.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	err = storage.WithRepo(deps.DBPath, func(repo *storage.SecretRepository) error {
		return funcs.NewExportSecretsFunc(repo, deps.Enc).Run(context.Background(), funcs.ExportSecretsDto{
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

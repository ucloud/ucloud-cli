package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPgsqlBackup ucloud pgsql backup
func newPgsqlBackup(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "List and manipulate backups of UPgSQL instances",
		Long:  "List and manipulate backups of UPgSQL instances",
	}
	cmd.AddCommand(newBackupList(ctx))
	cmd.AddCommand(newBackupDownload(ctx))
	cmd.AddCommand(newBackupStrategy(ctx))
	cmd.AddCommand(newBackupUpdateStrategy(ctx))
	return cmd
}

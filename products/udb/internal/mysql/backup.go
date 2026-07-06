package mysql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUDBBackup ucloud udb backup
func newUDBBackup(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "List and manipulate backups of MySQL instance",
		Long:  "List and manipulate backups of MySQL instance",
	}
	cmd.AddCommand(newUDBBackupCreate(ctx))
	cmd.AddCommand(newUDBBackupList(ctx))
	cmd.AddCommand(newUDBBackupDelete(ctx))
	cmd.AddCommand(newUDBBackupGetDownloadURL(ctx))
	return cmd
}

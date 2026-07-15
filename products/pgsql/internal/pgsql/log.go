package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPgsqlLog ucloud pgsql log
func newPgsqlLog(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "List and back up logs of UPgSQL instances",
		Long:  "List and back up logs of UPgSQL instances",
	}
	cmd.AddCommand(newLogList(ctx))
	cmd.AddCommand(newLogBackup(ctx))
	return cmd
}

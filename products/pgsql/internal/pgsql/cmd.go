package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `pgsql` root command and mounts the `db` subtree.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pgsql",
		Short: "Manipulate UPgSQL on UCloud platform",
		Long:  "Manipulate UPgSQL (UCloud PostgreSQL) on UCloud platform",
	}
	cmd.AddCommand(newPgsqlDB(ctx))
	cmd.AddCommand(newPgsqlConf(ctx))
	cmd.AddCommand(newPgsqlBackup(ctx))
	cmd.AddCommand(newPgsqlLog(ctx))
	return cmd
}

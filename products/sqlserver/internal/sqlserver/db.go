package sqlserver

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSQLServerDB builds the "sqlserver db" subcommand group for instance lifecycle operations.
func newSQLServerDB(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage SQL Server instances",
		Long:  "Manage SQL Server instances",
	}

	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	// cmd.AddCommand(newCreateAlwaysOn(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newStart(ctx))
	cmd.AddCommand(newStop(ctx))
	cmd.AddCommand(newRestart(ctx))

	return cmd
}

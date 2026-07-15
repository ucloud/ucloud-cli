package sqlserver

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `sqlserver` root command and mounts the `db` subtree.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sqlserver",
		Short: "Manipulate SQL Server on UCloud platform",
		Long:  "Manipulate SQL Server on UCloud platform",
	}
	cmd.AddCommand(newSQLServerDB(ctx))
	return cmd
}

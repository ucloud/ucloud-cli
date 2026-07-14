package udac

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `udac` root command and mounts the subcommands.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "udac",
		Short: "Manage Database Autonomous Center (UDAC) instances",
		Long:  "Import, export, and list database instances in the Database Autonomous Center (UDAC).",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newImport(ctx))
	cmd.AddCommand(newExport(ctx))
	cmd.AddCommand(newList(ctx))

	return cmd
}

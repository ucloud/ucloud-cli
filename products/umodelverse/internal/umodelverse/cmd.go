package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `umodelverse` root command.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   productName,
		Short: "Manipulate uModelVerse resources",
		Long:  "Manipulate uModelVerse resources, API keys, model catalog, inference logs, and billing data.",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newAPIKey(ctx))
	cmd.AddCommand(newModel(ctx))
	cmd.AddCommand(newLog(ctx))
	cmd.AddCommand(newOrder(ctx))
	cmd.AddCommand(newFilterOptions(ctx))
	return cmd
}

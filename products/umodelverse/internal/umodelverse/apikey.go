package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newAPIKey(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apikey",
		Short: "Manage uModelVerse API keys",
		Long:  "Manage uModelVerse API keys.",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newAPIKeyCreate(ctx))
	cmd.AddCommand(newAPIKeyDelete(ctx))
	cmd.AddCommand(newAPIKeyUpdate(ctx))
	cmd.AddCommand(newAPIKeyList(ctx))
	return cmd
}

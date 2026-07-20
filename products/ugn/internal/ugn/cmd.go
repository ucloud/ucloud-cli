package ugn

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `ugn` root command.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ugn",
		Short: "List and manipulate ugn instances",
		Long:  "List and manipulate ugn instances",
	}

	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newGet(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newRegion(ctx))
	cmd.AddCommand(newBW(ctx))
	cmd.AddCommand(newNetwork(ctx))
	cmd.AddCommand(newRoute(ctx))

	return cmd
}

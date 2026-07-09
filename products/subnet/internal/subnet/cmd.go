package subnet

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand returns the ucloud subnet command tree.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet",
		Short: "List, create and delete subnet",
		Long:  "List, create and delete subnet",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newListResource(ctx))
	return cmd
}

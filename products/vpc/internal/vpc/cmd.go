package vpc

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand returns the ucloud vpc command tree.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpc",
		Short: "List and manipulate VPC instances",
		Long:  "List and manipulate VPC instances",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newCreatePeer(ctx))
	cmd.AddCommand(newListPeer(ctx))
	cmd.AddCommand(newDeletePeer(ctx))
	return cmd
}

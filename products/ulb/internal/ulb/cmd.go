package ulb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand returns the ucloud ulb command tree.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ulb",
		Short: "List and manipulate ULB instances",
		Long:  "List and manipulate ULB instances",
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newUpdate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newVServer(ctx))
	cmd.AddCommand(newSSL(ctx))
	return cmd
}

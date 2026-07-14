package nlb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newListener assembles the `nlb listener` sub-tree.
func newListener(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listener",
		Short: "List and manipulate NLB listeners",
		Long:  "List and manipulate NLB listeners",
	}
	cmd.AddCommand(newListenerList(ctx))
	cmd.AddCommand(newListenerCreate(ctx))
	cmd.AddCommand(newListenerUpdate(ctx))
	cmd.AddCommand(newListenerDelete(ctx))
	return cmd
}

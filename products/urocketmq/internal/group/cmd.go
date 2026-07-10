package group

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `urocketmq group` resource-group command. Action subcommands are appended
// at the end in subsequent batches (order is fixed, golden depends).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage URocketMQ consumer groups",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newList(ctx))
	return cmd
}

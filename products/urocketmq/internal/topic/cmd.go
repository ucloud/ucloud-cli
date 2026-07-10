package topic

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `urocketmq topic` resource-group command. Action subcommands are appended
// at the end in subsequent batches (order is fixed, golden depends).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Manage URocketMQ topics",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newUpdate(ctx))
	return cmd
}

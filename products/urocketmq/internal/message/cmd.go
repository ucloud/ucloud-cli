package message

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `urocketmq message` resource-group command. Action subcommands are appended
// at the end in subsequent batches (order is fixed, golden depends).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "message",
		Short: "Query URocketMQ messages",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newQueryByID(ctx))
	cmd.AddCommand(newQueryByKey(ctx))
	cmd.AddCommand(newQueryByTopic(ctx))
	return cmd
}

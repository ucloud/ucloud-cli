package internal

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/group"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/message"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/token"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/topic"
)

// NewCommand builds the top-level urocketmq command.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "urocketmq",
		Short: "Manage URocketMQ instances, topics, groups, tokens and messages",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(service.NewCommand(ctx))
	cmd.AddCommand(topic.NewCommand(ctx))
	cmd.AddCommand(group.NewCommand(ctx))
	cmd.AddCommand(token.NewCommand(ctx))
	cmd.AddCommand(message.NewCommand(ctx))
	return cmd
}

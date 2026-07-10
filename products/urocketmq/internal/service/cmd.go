package service

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `urocketmq service` resource-group command. Action subcommands are appended
// at the end (order is fixed, golden depends);
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage URocketMQ service instances",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newGet(ctx))
	cmd.AddCommand(newUpdateName(ctx))
	cmd.AddCommand(newUpdateRemark(ctx))
	cmd.AddCommand(newPrice(ctx))
	return cmd
}

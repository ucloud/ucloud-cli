package token

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `urocketmq token` resource-group command. Action subcommand order is fixed as
// create/delete/get/list/update (golden depends, do not reorder arbitrarily).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage URocketMQ access tokens",
		Long:  "Create, delete, get, list and update URocketMQ access tokens for fine-grained topic access control.",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newGet(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newUpdate(ctx))
	return cmd
}

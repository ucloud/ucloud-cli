package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newOrder(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order",
		Short: "Query and export uModelVerse orders",
		Long:  "Query and export uModelVerse orders.",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newOrderAmount(ctx))
	cmd.AddCommand(newOrderPaid(ctx))
	cmd.AddCommand(newOrderUnpaid(ctx))
	cmd.AddCommand(newOrderSummary(ctx))
	return cmd
}

package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newOrderPaid(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paid",
		Short: "Query and export paid uModelVerse orders",
		Long:  "Query and export paid uModelVerse orders.",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newOrderPaidList(ctx))
	cmd.AddCommand(newOrderPaidSummary(ctx))
	cmd.AddCommand(newOrderPaidExport(ctx))
	return cmd
}

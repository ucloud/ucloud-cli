package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newOrderUnpaid(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpaid",
		Short: "Query and export unpaid uModelVerse orders",
		Long:  "Query and export unpaid uModelVerse orders.",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newOrderUnpaidList(ctx))
	cmd.AddCommand(newOrderUnpaidSummary(ctx))
	cmd.AddCommand(newOrderUnpaidExport(ctx))
	return cmd
}

package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newOrderSummary(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Export uModelVerse order summaries",
		Long:  "Export uModelVerse order summaries.",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newOrderSummaryExport(ctx))
	return cmd
}

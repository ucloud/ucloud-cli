package umodelverse

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newOrderPaidExport(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &orderRequest{}
	newRequest(client, req, false)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export paid uModelVerse order details",
		Long:  "Export paid uModelVerse order details as an Excel file download link.",
		Run: func(c *cobra.Command, args []string) {
			resp, err := invokeUMAction(client, "DownloadListPaidOrders", req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintln(ctx.ProgressWriter(), "umodelverse paid order export created")
			printResponse(ctx, resp)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	bindTimeRange(cmd, req)
	bindOrderFilters(cmd, req)
	return cmd
}

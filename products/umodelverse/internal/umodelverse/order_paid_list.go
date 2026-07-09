package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newOrderPaidList(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &orderRequest{}
	newRequest(client, req, true)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List paid uModelVerse orders",
		Long:  "List paid uModelVerse orders. Time range is [start-time, end-time), in Unix seconds.",
		Run: func(c *cobra.Command, args []string) {
			resp, err := invokeUMAction(client, "ListPaidOrders", req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			printResponse(ctx, resp)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	bindTimeRange(cmd, req)
	bindPage(cmd, req)
	bindOrderFilters(cmd, req)
	return cmd
}

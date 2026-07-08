package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newOrderUnpaidList(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &orderRequest{}
	newRequest(client, req, true)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List unpaid uModelVerse orders",
		Long:  "List unpaid uModelVerse orders.",
		Run: func(c *cobra.Command, args []string) {
			resp, err := invokeUMAction(client, "ListUnpaidOrders", req)
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

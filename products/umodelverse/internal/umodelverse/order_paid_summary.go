package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newOrderPaidSummary(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &orderRequest{}
	newRequest(client, req, true)

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Summarize paid uModelVerse orders",
		Long:  "Summarize paid uModelVerse orders.",
		Run: func(c *cobra.Command, args []string) {
			resp, err := invokeUMAction(client, "ListPaidOrderSummary", req)
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
	bindOrderFilters(cmd, req)
	bindOrderChargeTypes(cmd, req)
	return cmd
}

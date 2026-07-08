package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newOrderAmount(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &orderRequest{}
	newRequest(client, req, true)

	cmd := &cobra.Command{
		Use:   "amount",
		Short: "Get uModelVerse order amount statistics",
		Long:  "Get uModelVerse order amount statistics.",
		Run: func(c *cobra.Command, args []string) {
			resp, err := invokeUMAction(client, "GetOrderAmount", req)
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
	return cmd
}

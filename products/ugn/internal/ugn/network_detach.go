package ugn

import (
	"fmt"

	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newNetworkDetach ucloud ugn network detach
func newNetworkDetach(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewDetachUGNNetworksRequest()

	var networkIDs []string
	var yes bool

	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach network instances from ugn",
		Long:  "Detach network instances from ugn",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to detach the network instance(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			req.UGNID = sdk.String(ctx.PickResourceID(*req.UGNID))
			req.Networks = make([]string, 0, len(networkIDs))
			for _, id := range networkIDs {
				req.Networks = append(req.Networks, ctx.PickResourceID(id))
			}

			_, err = client.DetachUGNNetworks(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "networks detached from ugn[%s]\n", *req.UGNID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.UGNID, Action: "detach-network", Status: "Detached"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&networkIDs, "network-id", nil, "Required. Network IDs, e.g. vnet-xxxxx. Repeatable or comma-separated.")
	req.UGNID = flags.String("ugn-id", "", "Required. Resource ID of the ugn instance")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")

	cmd.MarkFlagRequired("network-id")
	cmd.MarkFlagRequired("ugn-id")

	return cmd
}

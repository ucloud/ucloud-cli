package ugn

import (
	"fmt"

	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newNetworkAttach ucloud ugn network attach
func newNetworkAttach(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewAttachUGNNetworksRequest()

	var networkIDs []string
	var networkOrgNames, networkRegions, networkTypes []string

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach network instances to ugn",
		Long:  "Attach network instances to ugn",
		Run: func(c *cobra.Command, args []string) {
			n := len(networkIDs)
			if len(networkTypes) != n || len(networkRegions) != n || len(networkOrgNames) != n {
				ctx.HandleError(fmt.Errorf(
					"network-id(%d), network-type(%d), network-region(%d), network-project-id(%d) must be provided in equal numbers",
					n, len(networkTypes), len(networkRegions), len(networkOrgNames)))
				return
			}
			networks := make([]ugnsdk.AttachUGNNetworksParamNetworks, 0, n)
			for i, id := range networkIDs {
				id = ctx.PickResourceID(id)
				networks = append(networks, ugnsdk.AttachUGNNetworksParamNetworks{
					NetworkID: sdk.String(id),
					OrgName:   sdk.String(networkOrgNames[i]),
					Region:    sdk.String(networkRegions[i]),
					Type:      sdk.String(networkTypes[i]),
				})
			}
			req.Networks = networks

			_, err := client.AttachUGNNetworks(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "networks attached to ugn[%s]\n", *req.UGNID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.UGNID, Action: "attach-network", Status: "Attached"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&networkIDs, "network-id", nil, "Required. Network IDs, e.g. vnet-xxxxx. Repeatable or comma-separated.")
	flags.StringSliceVar(&networkOrgNames, "network-project-id", nil, "Required. Project ID of the networks, one per network-id")
	flags.StringSliceVar(&networkRegions, "network-region", nil, "Required. Region of the networks, one per network-id")
	flags.StringSliceVar(&networkTypes, "network-type", nil, "Required. Network type, e.g. VPC/UCVR, one per network-id")
	req.UGNID = flags.String("ugn-id", "", "Required. Resource ID of the ugn instance")

	cmd.MarkFlagRequired("network-id")
	cmd.MarkFlagRequired("network-project-id")
	cmd.MarkFlagRequired("network-region")
	cmd.MarkFlagRequired("network-type")
	cmd.MarkFlagRequired("ugn-id")

	return cmd
}

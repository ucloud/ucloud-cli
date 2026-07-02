package eip

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newJoinSharedBW ucloud eip join-shared-bw
func newJoinSharedBW(ctx *cli.Context) *cobra.Command {
	eipIDs := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewAssociateEIPWithShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:     "join-shared-bw",
		Short:   "Join shared bandwidth",
		Long:    "Join shared bandwidth",
		Example: "ucloud eip join-shared-bw --eip-id eip-xxx --shared-bw-id bwshare-xxx",
		Run: func(c *cobra.Command, args []string) {
			for _, eip := range eipIDs {
				req.EIPIds = append(req.EIPIds, ctx.PickResourceID(eip))
			}
			req.ShareBandwidthId = sdk.String(ctx.PickResourceID(*req.ShareBandwidthId))
			_, err := client.AssociateEIPWithShareBandwidth(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "eip%v joined shared bandwidth[%s]\n", req.EIPIds, *req.ShareBandwidthId)
			results := []cli.OpResultRow{}
			for _, eipID := range req.EIPIds {
				results = append(results, cli.OpResultRow{ResourceID: eipID, Action: "join-shared-bw", Status: "Joined"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&eipIDs, "eip-id", nil, "Required. Resource ID of EIPs to join shared bandwdith")
	req.ShareBandwidthId = flags.String("shared-bw-id", "", "Required. Resource ID of shared bandwidth to be joined")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, nil, []string{EIP_CHARGE_BANDWIDTH, EIP_CHARGE_TRAFFIC})
	})
	command.SetCompletion(cmd, "shared-bw-id", func() []string {
		list, _ := getAllSharedBW(ctx, *req.ProjectId, *req.Region)
		return list
	})
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("shared-bw-id")

	return cmd
}

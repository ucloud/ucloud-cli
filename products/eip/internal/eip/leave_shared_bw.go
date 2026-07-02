package eip

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newLeaveSharedBW ucloud eip leave-shared-bw
func newLeaveSharedBW(ctx *cli.Context) *cobra.Command {
	eipIDs := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDisassociateEIPWithShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:     "leave-shared-bw",
		Short:   "Leave shared bandwidth",
		Long:    "Leave shared bandwidth",
		Example: "ucloud eip leave-shared-bw --eip-id eip-b2gvu3",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			if *req.ShareBandwidthId == "" {
				for _, eipID := range eipIDs {
					eipIns, err := getEIP(ctx, ctx.PickResourceID(eipID))
					if err != nil {
						ctx.HandleError(err)
						continue
					}
					sharedBWID := eipIns.ShareBandwidthSet.ShareBandwidthId
					if sharedBWID == "" {
						fmt.Fprintf(ctx.ProgressWriter(), "eip[%s] doesn't join any shared bandwidth\n", eipID)
						continue
					}
					req.ShareBandwidthId = sdk.String(sharedBWID)
					req.EIPIds = []string{ctx.PickResourceID(eipID)}
					_, err = client.DisassociateEIPWithShareBandwidth(req)
					if err != nil {
						ctx.HandleError(err)
						continue
					}
					fmt.Fprintf(ctx.ProgressWriter(), "eip[%s] left shared bandwidth[%s]\n", eipID, sharedBWID)
					results = append(results, cli.OpResultRow{ResourceID: ctx.PickResourceID(eipID), Action: "leave-shared-bw", Status: "Left"})
				}
			} else {
				for _, id := range eipIDs {
					req.EIPIds = append(req.EIPIds, ctx.PickResourceID(id))
				}
				*req.ShareBandwidthId = ctx.PickResourceID(*req.ShareBandwidthId)
				_, err := client.DisassociateEIPWithShareBandwidth(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "eip%v left shared bandwidth[%s]\n", eipIDs, *req.ShareBandwidthId)
				for _, eipID := range req.EIPIds {
					results = append(results, cli.OpResultRow{ResourceID: eipID, Action: "leave-shared-bw", Status: "Left"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&eipIDs, "eip-id", nil, "Required. Resource ID of EIPs to leave shared bandwidth")
	req.Bandwidth = flags.Int("bandwidth-mb", 1, "Required. Bandwidth of EIP after leaving shared bandwidth, ranging [1,300] for 'Traffic' charge mode, ranging [1,800] for 'Bandwidth' charge mode. Unit:Mb")
	req.PayMode = flags.String("traffic-mode", "Bandwidth", "Optional. Charge mode of the EIP after leaving shared bandwidth, 'Bandwidth' or 'Traffic'")
	req.ShareBandwidthId = flags.String("shared-bw-id", "", "Optional. Resource ID of shared bandwidth instance, assign this flag to make the operation faster")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	command.SetFlagValues(cmd, "traffic-mode", "Bandwidth", "Traffic")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, nil, []string{EIP_CHARGE_SHARE})
	})
	command.SetCompletion(cmd, "shared-bw-id", func() []string {
		list, _ := getAllSharedBW(ctx, *req.ProjectId, *req.Region)
		return list
	})

	// L2 prebug preserved verbatim: the flag is named "bandwidth-mb" (above), so
	// MarkFlagRequired("bandwidth") is a silent no-op. Matches cmd/eip.go ~:649.
	cmd.MarkFlagRequired("bandwidth")
	cmd.MarkFlagRequired("eip-id")
	return cmd
}

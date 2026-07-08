package bw

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newSharedResize returns ucloud bw shared resize.
func newSharedResize(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewResizeShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize shared bandwidth instance's bandwidth",
		Long:  "Resize shared bandwidth instance's bandwidth",
		Run: func(c *cobra.Command, args []string) {
			if *req.ShareBandwidth < 20 || *req.ShareBandwidth > 5000 {
				fmt.Fprintf(ctx.ProgressWriter(), "bandwidth should be between 20 and 5000. received %d\n", *req.ShareBandwidth)
				return
			}
			req.ShareBandwidthId = sdk.String(ctx.PickResourceID(*req.ShareBandwidthId))
			_, err := client.ResizeShareBandwidth(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "shared bandwidth[%s] resized to %dMb\n", *req.ShareBandwidthId, *req.ShareBandwidth)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.ShareBandwidthId, Action: "resize", Status: "Resized"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ShareBandwidthId = flags.String("shared-bw-id", "", "Required. Resource ID of shared bandwidth instance to resize")
	req.ShareBandwidth = flags.Int("bandwidth-mb", 0, "Required. Unit:Mb. resize to bandwidth value")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetCompletion(cmd, "shared-bw-id", func() []string {
		list, _ := getAllSharedBW(ctx, *req.ProjectId, *req.Region)
		return list
	})

	cmd.MarkFlagRequired("shared-bw-id")
	cmd.MarkFlagRequired("bandwidth-mb")

	return cmd
}

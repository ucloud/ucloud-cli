package bw

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newSharedCreate returns ucloud bw shared create.
func newSharedCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewAllocateShareBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create shared bandwidth instance",
		Long:  "Create shared bandwidth instance",
		Run: func(c *cobra.Command, args []string) {
			if *req.ShareBandwidth < 20 || *req.ShareBandwidth > 5000 {
				fmt.Fprintf(ctx.ProgressWriter(), "bandwidth should be between 20 and 5000. received %d\n", *req.ShareBandwidth)
				return
			}
			resp, err := client.AllocateShareBandwidth(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "shared bandwidth[%s] created\n", resp.ShareBandwidthId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.ShareBandwidthId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Name = flags.String("name", "", "Required. Name of the shared bandwidth instance")
	req.ShareBandwidth = flags.Int("bandwidth-mb", 20, "Optional. Unit:Mb. Bandwidth of the shared bandwidth. Range [20,5000]")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic")

	cmd.MarkFlagRequired("name")

	return cmd
}

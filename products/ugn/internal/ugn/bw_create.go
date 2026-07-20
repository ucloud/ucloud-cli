package ugn

import (
	"fmt"

	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBWCreate ucloud ugn bw create
func newBWCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewCreateSimpleUGNBwPackageRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a ugn bandwidth package",
		Long:  "Create a ugn bandwidth package",
		Run: func(c *cobra.Command, args []string) {
			_, err := client.CreateSimpleUGNBwPackage(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "ugn bw created\n")
			ctx.EmitResult(cli.OpResultRow{Action: "create-bw", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.UGNID = flags.String("ugn-id", "", "Required. Resource ID of the ugn instance")
	req.BandWidth = flags.Int("bandwidth", 0, "Required. Bandwidth value in Mbps")
	req.Path = flags.String("path", "IGP", "Path policy: Delay/IGP/TCO")
	req.PayMode = flags.String("pay-mode", "", "Required. Pay mode: FixedBw/Max5/Traffic")
	req.ChargeType = flags.String("charge-type", "", "Required. Charge type: Month/Postpay")
	req.RegionA = flags.String("region-a", "", "Required. Region A of the bandwidth package")
	req.RegionB = flags.String("region-b", "", "Required. Region B of the bandwidth package")
	req.Name = flags.String("name", "", "Optional. Bandwidth package name")
	req.Qos = flags.String("qos", "Platinum", "Optional. QoS: Diamond/Platinum/Gold")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID")
	req.Quantity = flags.Float64("quantity", 1, "Optional. Duration in months, default 1")

	ctx.BindProjectID(cmd, req)
	ctx.SetCompletion(cmd, "ugn-id", func() []string {
		return getAllUGNIdNames(ctx, *req.ProjectId)
	})
	ctx.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetFlagValues(cmd, "path", "Delay", "IGP", "TCO")
	command.SetFlagValues(cmd, "pay-mode", "FixedBw", "Max5", "Traffic")
	command.SetFlagValues(cmd, "charge-type", "Month", "Postpay")
	command.SetFlagValues(cmd, "qos", "Diamond", "Platinum", "Gold")

	cmd.MarkFlagRequired("ugn-id")
	cmd.MarkFlagRequired("bandwidth")
	cmd.MarkFlagRequired("pay-mode")
	cmd.MarkFlagRequired("charge-type")
	cmd.MarkFlagRequired("region-a")
	cmd.MarkFlagRequired("region-b")

	return cmd
}

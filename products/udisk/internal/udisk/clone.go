package udisk

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newClone ucloud udisk clone
func newClone(ctx *cli.Context) *cobra.Command {
	var async *bool
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewCloneUDiskRequest()
	enableDataArk := sdk.String("false")
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone an udisk",
		Long:  "Clone an udisk",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			if *enableDataArk == "true" {
				req.UDataArkMode = sdk.String("Yes")
			} else {
				req.UDataArkMode = sdk.String("No")
			}
			if strings.Index(*req.SourceId, "/") > -1 {
				*req.SourceId = strings.SplitN(*req.SourceId, "/", 2)[0]
			}
			resp, err := client.CloneUDisk(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.UDiskId) == 1 {
				text := fmt.Sprintf("cloned udisk:[%s] is initializing", resp.UDiskId[0])
				if *async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeUdiskByID(ctx)).Spoll(resp.UDiskId[0], text, []string{DISK_AVAILABLE, DISK_FAILED})
				}
				ctx.EmitResult(cli.OpResultRow{ResourceID: resp.UDiskId[0], Action: "clone", Status: "Initializing"})
			} else {
				fmt.Fprintf(w, "udisk[%v] cloned", resp.UDiskId)
				results := []cli.OpResultRow{}
				for _, id := range resp.UDiskId {
					results = append(results, cli.OpResultRow{ResourceID: id, Action: "clone", Status: "Cloned"})
				}
				ctx.EmitResult(results...)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.SourceId = flags.String("source-id", "", "Required. Resource ID of parent udisk")
	req.Name = flags.String("name", "", "Required. Name of new udisk")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	enableDataArk = flags.String("enable-data-ark", "false", "Optional. DataArk supports real-time backup, which can restore the udisk back to any moment within the last 12 hours.")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID, The Coupon can deduct part of the payment,see https://accountv2.ucloud.cn")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "enable-data-ark", "true", "false")

	command.SetCompletion(cmd, "source-id", func() []string {
		return getDiskList(ctx, []string{DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("source-id")
	cmd.MarkFlagRequired("name")

	return cmd
}

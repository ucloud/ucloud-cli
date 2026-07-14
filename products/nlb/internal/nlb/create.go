package nlb

import (
	"fmt"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate implements `nlb create`.
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewCreateNetworkLoadBalancerRequest()

	var couponID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an NLB instance",
		Long:  "Create an NLB (Network Load Balancer) instance in the specified VPC and subnet.",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.SubnetId = sdk.String(ctx.PickResourceID(*req.SubnetId))
			if c.Flags().Changed("coupon-id") {
				req.CouponId = &couponID
			}

			resp, err := client.CreateNetworkLoadBalancer(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "nlb[%s] created\n", resp.NLBId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.NLBId, Action: "create", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	req.VPCId = flags.String("vpc-id", "", "Required. Resource ID of the VPC the NLB belongs to. See 'ucloud vpc list'.")
	req.SubnetId = flags.String("subnet-id", "", "Required. Resource ID of the subnet the NLB belongs to. See 'ucloud subnet list'.")
	req.Name = flags.String("name", "", "Optional. NLB instance name, 1-255 chars.")
	req.IPVersion = flags.String("ip-version", "IPv4", "Optional. IP protocol version: IPv4/IPv6/DualStack.")
	req.ChargeType = flags.String("charge-type", "Dynamic", "Optional. Charge type: Dynamic (by hour), Month, Year.")
	req.Quantity = flags.Int("quantity", 1, "Optional. Purchase duration. For Month with value 0 means until end of month.")
	req.Tag = flags.String("group", "Default", "Optional. Business group.")
	req.Remark = flags.String("remark", "", "Optional. Remark of the NLB instance.")
	flags.StringVar(&couponID, "coupon-id", "", "Optional. Coupon ID.")

	command.SetFlagValues(cmd, "ip-version", "IPv4", "IPv6", "DualStack")
	command.SetFlagValues(cmd, "charge-type", "Dynamic", "Month", "Year")
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, derefStr(req.VPCId), derefStr(req.ProjectId), derefStr(req.Region))
	})

	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("subnet-id")

	return cmd
}

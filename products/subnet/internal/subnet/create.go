package subnet

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate returns ucloud subnet create.
func newCreate(ctx *cli.Context) *cobra.Command {
	var segment *net.IPNet
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewCreateSubnetRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create subnet of vpc network",
		Long:    "Create subnet of vpc network",
		Example: "ucloud subnet create --vpc-id uvnet-vpcxid --name testName --segment 192.168.2.0/24",
		Run: func(cmd *cobra.Command, args []string) {
			ipMaskStrs := strings.SplitN(segment.String(), "/", 2)
			req.Subnet = sdk.String(ipMaskStrs[0])
			mask, err := strconv.Atoi(ipMaskStrs[1])
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.Netmask = sdk.Int(mask)
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			resp, err := client.CreateSubnet(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "subnet[%s] created\n", resp.SubnetId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.SubnetId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.VPCId = flags.String("vpc-id", "", "Required. Assign the VPC network of the subnet")
	segment = flags.IPNet("segment", net.IPNet{}, "Required. Segment of subnet. For example '192.168.0.0/24'")
	req.SubnetName = flags.String("name", "Subnet", "Optional. Name of subnet to create")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Tag = flags.String("group", "", "Optional. Business group")
	req.Remark = flags.String("remark", "", "Optional. Remark of subnet to create")

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("segment")

	return cmd
}

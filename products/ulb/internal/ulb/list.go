package ulb

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList returns ucloud ulb list.
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeULBRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ULB instances",
		Long:  "List ULB instances",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			resp, err := client.DescribeULB(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []Row{}
			for _, ulb := range resp.DataSet {
				row := Row{}
				row.ResourceID = ulb.ULBId
				row.Name = ulb.Name
				row.Group = ulb.BusinessId
				row.VserverCount = len(ulb.VServerSet)
				row.VPC = ulb.VPCId
				row.CreationTime = common.FormatDate(ulb.CreateTime)
				if ulb.ULBType == "OuterMode" {
					ips := []string{}
					for _, ip := range ulb.IPSet {
						ips = append(ips, fmt.Sprintf("%s(%s)", ip.EIP, ip.EIPId))
					}
					row.Network = strings.Join(ips, ",")
				} else {
					row.Network = ulb.PrivateIP
				}
				list = append(list, row)
			}

			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	req.ULBId = flags.String("ulb-id", "", "Optional. Resource ID of ULB instance to list")
	req.VPCId = flags.String("vpc-id", "", "Optional. Resource ID of VPC which the ULB instances to list belong to")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Resource ID of subnet which the ULB instances to list belong to")
	req.BusinessId = flags.String("group", "", "Optional. Business group of ULB instances to list")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

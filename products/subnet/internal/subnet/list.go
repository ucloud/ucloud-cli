package subnet

import (
	"fmt"

	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList returns ucloud subnet list.
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDescribeSubnetRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subnet",
		Long:  `List subnet`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeSubnet(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := make([]Row, 0)
			for _, sn := range resp.DataSet {
				row := Row{}
				row.SubnetName = sn.SubnetName
				row.ResourceID = sn.SubnetId
				row.Group = sn.Tag
				row.AffiliatedVPC = fmt.Sprintf("%s/%s", sn.VPCId, sn.VPCName)
				row.NetworkSegment = fmt.Sprintf("%s/%s", sn.Subnet, sn.Netmask)
				row.CreationTime = common.FormatDate(sn.CreateTime)
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringSliceVar(&req.SubnetIds, "subnet-id", []string{}, "Optional. Multiple values separated by commas")
	req.VPCId = flags.String("vpc-id", "", "Optional. Resource ID of VPC")
	req.Tag = flags.String("group", "", "Optional. Group")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")

	return cmd
}

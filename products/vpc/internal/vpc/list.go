package vpc

import (
	"strings"

	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList returns ucloud vpc list.
func newList(ctx *cli.Context) *cobra.Command {
	vpcIDs := []string{}
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDescribeVPCRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List vpc",
		Long:  "List vpc",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			req.VPCIds = nil
			for _, id := range vpcIDs {
				req.VPCIds = append(req.VPCIds, ctx.PickResourceID(id))
			}
			resp, err := client.DescribeVPC(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []Row{}
			for _, vpc := range resp.DataSet {
				row := Row{}
				row.VPCName = vpc.Name
				row.ResourceID = vpc.VPCId
				row.Group = vpc.Tag
				row.NetworkSegment = strings.Join(vpc.Network, ",")
				row.SubnetCount = vpc.SubnetCount
				row.CreationTime = common.FormatDate(vpc.CreateTime)
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Tag = flags.String("group", "", "Optional. Group")
	flags.StringSliceVar(&vpcIDs, "vpc-id", []string{}, "Optional. Multiple values separated by commas")

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

package vpc

import (
	"strings"

	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListPeer returns ucloud vpc list-intercome.
func newListPeer(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDescribeVPCIntercomRequest()
	cmd := &cobra.Command{
		Use:     "list-intercome",
		Short:   "list intercome ",
		Long:    "list intercome",
		Example: "ucloud vpc list-intercome --vpc-id xx",
		Run: func(cmd *cobra.Command, args []string) {
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			resp, err := client.DescribeVPCIntercom(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := make([]IntercomRow, 0)
			for _, vpcIntercom := range resp.DataSet {
				row := IntercomRow{}
				row.ProjectID = vpcIntercom.ProjectId
				row.Segments = strings.Join(vpcIntercom.Network, ",")
				row.DstRegion = vpcIntercom.DstRegion
				row.VPCName = vpcIntercom.Name
				row.ResourceID = vpcIntercom.VPCId
				row.Group = vpcIntercom.Tag
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	req.VPCId = cmd.Flags().String("vpc-id", "", "Required. The vpc id which you wnat to describe the information")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. The project id of source vpc")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional, The region of source vpc")

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)

	cmd.MarkFlagRequired("vpc-id")

	return cmd
}

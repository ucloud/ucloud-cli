package subnet

import (
	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListResource returns ucloud subnet list-resource.
func newListResource(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDescribeSubnetResourceRequest()
	cmd := &cobra.Command{
		Use:   "list-resource",
		Short: "List resources belong to subnet",
		Long:  "List resources belong to subnet",
		Run: func(c *cobra.Command, args []string) {
			req.SubnetId = sdk.String(ctx.PickResourceID(*req.SubnetId))
			resp, err := client.DescribeSubnetResource(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []ResourceRow{}
			for _, r := range resp.DataSet {
				row := ResourceRow{
					ResourceName: r.Name,
					ResourceID:   r.ResourceId,
					ResourceType: r.ResourceType,
					PrivateIP:    r.IP,
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.SubnetId = flags.String("subnet-id", "", "Required. Resource ID of subnet which resources to list belong to")
	req.ResourceType = flags.String("resource-type", "", "Optional. Resource type of resources to list. Accept values:'uhost','phost','ulb','uhadoophost','ufortresshost','unatgw','ukafka','umem','docker','udb','udw' and 'vip'")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindLimit(cmd, req)
	ctx.BindOffset(cmd, req)
	cmd.MarkFlagRequired("subnet-id")
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, "", *req.ProjectId, *req.Region)
	})
	command.SetFlagValues(cmd, "resource-type", "uhost", "phost", "ulb", "uhadoophost", "ufortresshost", "unatgw", "ukafka", "umem", "docker", "udb", "udw", "vip")

	return cmd
}

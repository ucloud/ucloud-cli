package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBackendList returns ucloud ulb vserver backend list.
func newBackendList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeVServerRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ULB VServer backend nodes",
		Long:  "List ULB VServer backend nodes",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(ctx.PickResourceID(*req.VServerId))
			resp, err := client.DescribeVServer(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.DataSet) != 1 {
				fmt.Fprintf(ctx.ProgressWriter(), "ulb[%s] or vserver[%s] may not exist\n", *req.ULBId, *req.VServerId)
				return
			}
			vs := resp.DataSet[0]
			list := []BackendRow{}
			for _, node := range vs.BackendSet {
				row := BackendRow{}
				row.Name = node.ResourceName
				row.ResourceID = node.ResourceId
				row.BackendID = node.BackendId
				row.PrivateIP = node.PrivateIP
				row.Weight = node.Weight
				row.Port = node.Port
				if node.Status == 0 {
					row.HealthCheck = "Normal"
				} else if node.Status == 1 {
					row.HealthCheck = "Failed"
				}
				if node.Enabled == 1 {
					row.NodeMode = "enable"
				} else if node.Enabled == 0 {
					row.NodeMode = "disable"
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB which the backend nodes belong to")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer which the backend nodes belong to")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")

	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "vserver-id", func() []string {
		ulbID := ctx.PickResourceID(*req.ULBId)
		return getAllVServerIDNames(ctx, ulbID, *req.ProjectId, *req.Region)
	})

	return cmd
}

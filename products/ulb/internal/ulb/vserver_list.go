package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newVServerList returns ucloud ulb vserver list.
func newVServerList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeVServerRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ULB Vserver instances",
		Long:  "List ULB Vserver instances",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			resp, err := client.DescribeVServer(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []VServerRow{}
			for _, vs := range resp.DataSet {
				row := VServerRow{}
				row.VServerName = vs.VServerName
				row.ResourceID = vs.VServerId
				row.ListenType = vs.ListenType
				row.Protocol = vs.Protocol
				row.Port = vs.FrontendPort
				row.LBMethod = vs.Method
				row.ClientTimeout = fmt.Sprintf("%ds", vs.ClientTimeout)
				row.SessionMaintainMode = vs.PersistenceType
				row.SessionMaintainKey = vs.PersistenceInfo
				row.HealthCheckMode = vs.MonitorType
				row.HealthCheckDomain = vs.Domain
				row.HealthCheckPath = vs.Path
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Optional. Resource ID of vserver to list")

	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")

	return cmd
}

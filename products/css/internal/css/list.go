package css

import (
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList ucloud css list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewListUESInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UES instances",
		Long:  "List UES instances",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.ListUESInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []InstanceRow{}
			for _, ins := range resp.ClusterSet {
				row := InstanceRow{
					InstanceID:   ins.InstanceId,
					InstanceName: ins.InstanceName,
					AppName:      ins.AppName,
					AppVersion:   ins.AppVersion,
					Zone:         ins.Zone,
					State:        ins.State,
					NodeCount:    fmt.Sprintf("%d", ins.NodeCount),
					VPCId:        ins.VPCId,
					SubnetId:     ins.SubnetId,
					ChargeType:   ins.ChargeType,
					CreateTime:   common.FormatDate(ins.CreateTime),
					ExpireTime:   common.FormatDate(ins.ExpireTime),
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", "", "Optional. Assign availability zone")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 30, "Optional. Limit, default 30")
	return cmd
}

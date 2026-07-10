package group

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newList ucloud urocketmq group list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewListURocketMQGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List URocketMQ consumer groups",
		Long:  "List URocketMQ consumer groups",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.ListURocketMQGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			listGroup(ctx, resp.GroupList)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit default 50")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")

	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("service-id")

	return cmd
}

// listGroup renders the group list. json/yaml emits full-field groupRow; table mode uses groupRowDefault.
func listGroup(ctx *cli.Context, groups []urocketmq.GroupBaseInfo) {
	list := make([]groupRow, 0, len(groups))
	for _, g := range groups {
		list = append(list, groupRow{
			GroupName:  g.GroupName,
			Id:         g.Id,
			Remark:     g.Remark,
			CreateTime: g.CreateTime,
		})
	}

	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(list)
		return
	}

	rows := make([]groupRowDefault, 0, len(list))
	for _, r := range list {
		rows = append(rows, groupRowDefault{
			GroupName:  r.GroupName,
			Id:         r.Id,
			Remark:     r.Remark,
			CreateTime: common.FormatDate(r.CreateTime),
		})
	}
	ctx.PrintList(rows)
}

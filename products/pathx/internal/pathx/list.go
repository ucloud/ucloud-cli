package pathx

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList ucloud pathx list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	getPathxListReq := client.NewDescribeUGA3InstanceRequest()
	var instanceId string
	var detail bool
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all the pathx resource of project",
		Long:    "List all the pathx resource of project",
		Example: "'ucloud pathx list or ucloud pathx list --id uga-xxx or ucloud pathx list --id uga-xxx --detail",
		Run: func(cmd *cobra.Command, args []string) {
			if len(instanceId) > 0 {
				getPathxListReq.InstanceId = &instanceId
			}
			resp, err := client.DescribeUGA3Instance(getPathxListReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			forwardInfos := resp.ForwardInstanceInfos
			if len(forwardInfos) == 0 {
				ctx.HandleError(fmt.Errorf("No pathx resource found under the current project."))
				return
			}
			if detail && len(instanceId) > 0 {
				printPathxDetail(ctx, forwardInfos[0], ctx.ProgressWriter())
				return
			}
			list := make([]Uga3DescribeRow, 0)
			for _, item := range forwardInfos {
				egressIps := []string{}
				for _, egressIp := range item.EgressIpList {
					egressIps = append(egressIps, fmt.Sprintf("%s:%s", egressIp.Area, egressIp.IP))
				}
				list = append(list, Uga3DescribeRow{
					ResourceID:       item.InstanceId,
					CName:            item.CName,
					Name:             item.Name,
					AccelerationArea: item.AccelerationArea,
					Bandwidth:        item.Bandwidth,
					OriginAreaCode:   item.OriginAreaCode,
					IPList:           strings.Join(item.IPList, ","),
					Domain:           item.Domain,
					CreateTime:       common.FormatDate(item.CreateTime),
					EgressIpList:     strings.Join(egressIps, "|"),
				})
			}
			ctx.PrintList(list)
		},
	}
	flags := listCmd.Flags()
	flags.SortFlags = false
	ctx.BindProjectID(listCmd, getPathxListReq)
	ctx.BindRegion(listCmd, getPathxListReq)
	ctx.BindZone(listCmd, getPathxListReq)
	flags.StringVar(&instanceId, "id", "", "Required. It is the resource ID of pathx resource")
	flags.BoolVar(&detail, "detail", false, "Optional. If it is specified,the details will be printed")
	ctx.SetCompletion(listCmd, "id", func() []string {
		return getPathxList(ctx, *getPathxListReq.ProjectId, *getPathxListReq.Region, *getPathxListReq.Zone)
	})
	return listCmd
}

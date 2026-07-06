package uhost

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newIsolationList ucloud uhost isolation-group list
func newIsolationList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List isolation group of uhost",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeIsolationGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			var list []isolationGroupRow
			for _, group := range resp.IsolationGroupSet {
				row := isolationGroupRow{
					ResourceID: group.GroupId,
					Name:       group.GroupName,
					Remark:     group.Remark,
				}
				var zones []string
				for _, item := range group.SpreadInfoSet {
					zones = append(zones, fmt.Sprintf("%s:%d", item.Zone, item.UHostCount))
				}
				row.UHostCount = strings.Join(zones, " ")
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.GroupId = flags.String("group-id", "", "Optional. Resource ID of isolation group to describe")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindLimit(cmd, req)
	ctx.BindOffset(cmd, req)

	command.SetCompletion(cmd, "group-id", func() []string {
		return getIsolationGroupList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

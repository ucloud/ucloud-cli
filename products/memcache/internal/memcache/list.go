package memcache

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList returns ucloud memcache list.
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDescribeUMemcacheGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List memcache instances",
		Long:  "List memcache instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeUMemcacheGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []Row{}
			for _, ins := range resp.DataSet {
				row := Row{
					ResourceID: ins.GroupId,
					Name:       ins.Name,
					Group:      ins.Tag,
					Size:       fmt.Sprintf("%dGB", ins.Size),
					UsedSize:   fmt.Sprintf("%dMB", ins.UsedSize),
					State:      ins.State,
					CreateTime: common.FormatDate(ins.CreateTime),
					Address:    fmt.Sprintf("%s:%d", ins.VirtualIP, ins.Port),
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.GroupId = flags.String("umem-id", "", "Optional. Resource ID of the redis to list")
	ctx.BindRegion(cmd, req)
	ctx.BindZoneEmpty(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindOffset(cmd, req)
	ctx.BindLimit(cmd, req)

	return cmd
}

package redis

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList returns ucloud redis list.
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDescribeUMemRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List redis instances",
		Long:  "List redis instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeUMem(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []Row{}
			for _, ins := range resp.DataSet {
				row := Row{
					ResourceID: ins.ResourceId,
					Name:       ins.Name,
					Role:       ins.Role,
					Type:       string(resourceTypeToMode(ins.ResourceType)),
					Group:      ins.Tag,
					Size:       fmt.Sprintf("%dGB", ins.Size),
					UsedSize:   fmt.Sprintf("%dMB", ins.UsedSize),
					State:      ins.State,
					Zone:       ins.Zone,
					CreateTime: common.FormatDate(ins.CreateTime),
				}
				addrs := []string{}
				for _, addr := range ins.Address {
					addrs = append(addrs, fmt.Sprintf("%s:%d", addr.IP, addr.Port))
				}
				row.Address = strings.Join(addrs, "|")
				list = append(list, row)
				for _, slave := range ins.DataSet {
					srow := Row{
						ResourceID: slave.GroupId,
						Name:       slave.Name,
						Role:       fmt.Sprintf("⮑ %s", slave.Role),
						Type:       string(resourceTypeToMode(slave.ResourceType)),
						Group:      slave.Tag,
						Size:       fmt.Sprintf("%dGB", slave.Size),
						UsedSize:   fmt.Sprintf("%dMB", slave.UsedSize),
						State:      slave.State,
						Zone:       slave.Zone,
						Address:    fmt.Sprintf("%s:%d", slave.VirtualIP, slave.Port),
						CreateTime: common.FormatDate(slave.CreateTime),
					}
					list = append(list, srow)
				}
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ResourceId = flags.String("umem-id", "", "Optional. Resource ID of the redis to list")
	ctx.BindRegion(cmd, req)
	ctx.BindZoneEmpty(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindOffset(cmd, req)
	ctx.BindLimit(cmd, req)
	req.Protocol = sdk.String("redis")

	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

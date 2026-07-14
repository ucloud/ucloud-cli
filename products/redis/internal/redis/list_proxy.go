package redis

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListProxy returns ucloud redis list-proxy.
func newListProxy(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDescribeUDRedisProxyInfoRequest()
	cmd := &cobra.Command{
		Use:   "list-proxy",
		Short: "List proxy info of distributed redis",
		Long:  "List proxy info of distributed redis",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeUDRedisProxyInfo(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []ProxyRow{}
			for _, p := range resp.DataSet {
				row := ProxyRow{
					ProxyID:    p.ProxyId,
					ResourceID: p.ResourceId,
					State:      p.State,
					Vip:        p.Vip,
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.SpaceId = flags.String("umem-id", "", "Required. Resource ID of the distributed redis")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("umem-id")
	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

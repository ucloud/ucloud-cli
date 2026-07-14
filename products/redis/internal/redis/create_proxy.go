package redis

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreateProxy returns ucloud redis create-proxy.
func newCreateProxy(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewCreateUDRedisUhproxyRequest()
	cmd := &cobra.Command{
		Use:     "create-proxy",
		Short:   "Create proxy for distributed redis",
		Long:    "Create proxy for distributed redis",
		Example: "ucloud redis create-proxy --umem-id udb-xxx --cpu 2",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.CreateUDRedisUhproxy(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "proxy[%s] created\n", resp.ResourceId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.ResourceId, Action: "create-proxy", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.SpaceId = flags.String("umem-id", "", "Required. Resource ID of the distributed redis")
	req.CPU = flags.Int("cpu", 2, "Required. CPU cores of the proxy")
	req.Port = flags.Int("port", 6379, "Optional. Port of the proxy. Default value 6379")
	req.ProxyCnt = flags.Int("proxy-cnt", 1, "Optional. Number of proxies to create. Default value 1")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("umem-id")
	cmd.MarkFlagRequired("cpu")
	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

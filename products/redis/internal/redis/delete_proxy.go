package redis

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDeleteProxy returns ucloud redis delete-proxy.
func newDeleteProxy(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDeleteUDRedisProxyRequest()
	cmd := &cobra.Command{
		Use:     "delete-proxy",
		Short:   "Delete proxy of distributed redis",
		Long:    "Delete proxy of distributed redis",
		Example: "ucloud redis delete-proxy --umem-id udb-xxx --proxy-id proxy-xxx",
		Run: func(c *cobra.Command, args []string) {
			_, err := client.DeleteUDRedisProxy(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "proxy[%s] deleted\n", *req.ProxyId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.ProxyId, Action: "delete-proxy", Status: "Deleted"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.SpaceId = flags.String("umem-id", "", "Required. Resource ID of the distributed redis")
	req.ProxyId = flags.String("proxy-id", "", "Required. Proxy ID of the proxy to delete")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("umem-id")
	cmd.MarkFlagRequired("proxy-id")
	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

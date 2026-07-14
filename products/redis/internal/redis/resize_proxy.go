package redis

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newResizeProxy returns ucloud redis resize-proxy.
func newResizeProxy(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewResizeUhproxyRequest()
	cmd := &cobra.Command{
		Use:     "resize-proxy",
		Short:   "Resize proxy of distributed redis",
		Long:    "Resize proxy of distributed redis",
		Example: "ucloud redis resize-proxy --umem-id udb-xxx --proxy-id proxy-xxx --new-cpu 4",
		Run: func(c *cobra.Command, args []string) {
			_, err := client.ResizeUhproxy(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "proxy[%s] resized\n", *req.ProxyId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.ProxyId, Action: "resize-proxy", Status: "Resized"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.SpaceId = flags.String("umem-id", "", "Required. Resource ID of the distributed redis")
	req.ProxyId = flags.String("proxy-id", "", "Required. Proxy ID of the proxy to resize")
	req.NewCPU = flags.Int("new-cpu", 0, "Required. Target CPU cores of the proxy")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("umem-id")
	cmd.MarkFlagRequired("proxy-id")
	cmd.MarkFlagRequired("new-cpu")
	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

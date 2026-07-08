package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newSSLBind returns ucloud ulb ssl bind.
func newSSLBind(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewBindSSLRequest()
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Bind SSL Certificate with VServer",
		Long:  "Bind SSL Certificate with VServer",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(ctx.PickResourceID(*req.VServerId))
			req.SSLId = sdk.String(ctx.PickResourceID(*req.SSLId))
			_, err := client.BindSSL(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "ssl certificate[%s] bind with vserver[%s] of ulb[%s]\n", *req.SSLId, *req.VServerId, *req.ULBId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.SSLId, Action: "bind-ssl", Status: "Bound"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.SSLId = flags.String("ssl-id", "", "Required. Resource ID of SSL Certificate to bind")
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer")
	command.SetCompletion(cmd, "ssl-id", func() []string {
		return getAllSSLCertIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "vserver-id", func() []string {
		return getAllVServerIDNames(ctx, *req.ULBId, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ssl-id")
	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	return cmd
}

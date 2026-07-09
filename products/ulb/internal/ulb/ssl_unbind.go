package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newSSLUnbind returns ucloud ulb ssl unbind.
func newSSLUnbind(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewUnbindSSLRequest()
	cmd := &cobra.Command{
		Use:   "unbind",
		Short: "Unbind SSL Certificate with VServer",
		Long:  "Unbind SSL Certificate with VServer",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			req.VServerId = sdk.String(ctx.PickResourceID(*req.VServerId))
			req.SSLId = sdk.String(ctx.PickResourceID(*req.SSLId))
			_, err := client.UnbindSSL(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "ssl certificate[%s] unbind with vserver[%s] of ulb[%s]\n", *req.SSLId, *req.VServerId, *req.ULBId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.SSLId, Action: "unbind-ssl", Status: "Unbound"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.SSLId = flags.String("ssl-id", "", "Required. Resource ID of SSL Certificate to unbind")
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB")
	req.VServerId = flags.String("vserver-id", "", "Required. Resource ID of VServer")
	command.SetCompletion(cmd, "ssl-id", func() []string {
		return getAllSSLCertIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "ulb-id", func() []string {
		if *req.SSLId == "" {
			return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
		}
		cert, err := getSSLCertByID(ctx, *req.SSLId, *req.ProjectId, *req.Region)
		if err != nil {
			return nil
		}
		ulbs := []string{}
		for _, b := range cert.BindedTargetSet {
			ulbs = append(ulbs, fmt.Sprintf("%s/%s", b.ULBId, b.ULBName))
		}
		return ulbs
	})
	command.SetCompletion(cmd, "vserver-id", func() []string {
		if *req.SSLId == "" {
			return getAllVServerIDNames(ctx, *req.ULBId, *req.ProjectId, *req.Region)
		}
		cert, err := getSSLCertByID(ctx, *req.SSLId, *req.ProjectId, *req.Region)
		if err != nil {
			return nil
		}
		vservers := []string{}
		for _, b := range cert.BindedTargetSet {
			vservers = append(vservers, fmt.Sprintf("%s/%s", b.VServerId, b.VServerName))
		}
		return vservers
	})
	cmd.MarkFlagRequired("ssl-id")
	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")
	return cmd
}

package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newSSLDelete returns ucloud ulb ssl delete.
func newSSLDelete(ctx *cli.Context) *cobra.Command {
	var idNames []string
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDeleteSSLRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete SSL Certificates by resource id(ssl id)",
		Long:  "Delete SSL Certificates by resource id(ssl id)",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.SSLId = sdk.String(id)
				_, err := client.DeleteSSL(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "ssl certificate[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete-ssl", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringSliceVar(&idNames, "ssl-id", nil, "Required. Resource ID of SSL Certificates to delete")
	command.SetCompletion(cmd, "ssl-id", func() []string {
		return getAllSSLCertIDNames(ctx, *req.ProjectId, *req.Region)
	})
	return cmd
}

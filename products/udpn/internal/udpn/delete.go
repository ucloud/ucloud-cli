package udpn

import (
	"fmt"

	"github.com/spf13/cobra"

	udpnsdk "github.com/ucloud/ucloud-sdk-go/services/udpn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDelete ucloud udpn delete
func newDelete(ctx *cli.Context) *cobra.Command {
	idNames := []string{}
	client := cli.NewServiceClient(ctx, udpnsdk.NewClient)
	req := client.NewReleaseUDPNRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete udpn instances",
		Long:  "delete udpn instances",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.UDPNId = sdk.String(id)
				_, err := client.ReleaseUDPN(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "udpn[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udpn-id", nil, "Required. Resource ID of udpn instances to delete")
	ctx.BindProjectID(cmd, req)

	ctx.SetCompletion(cmd, "udpn-id", func() []string {
		return getAllUDPNIdNames(ctx, *req.ProjectId, ctx.DefaultRegion())
	})

	cmd.MarkFlagRequired("udpn-id")

	return cmd
}

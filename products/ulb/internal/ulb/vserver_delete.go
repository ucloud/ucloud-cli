package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newVServerDelete returns ucloud ulb vserver delete.
func newVServerDelete(ctx *cli.Context) *cobra.Command {
	vserverIDs := []string{}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDeleteVServerRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete ULB VServer instances",
		Long:  "Delete ULB VServer instances",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			results := []cli.OpResultRow{}
			for _, idname := range vserverIDs {
				vsid := ctx.PickResourceID(idname)
				req.VServerId = sdk.String(vsid)
				_, err := client.DeleteVServer(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "ulb-vserver[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: vsid, Action: "delete-vserver", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB instance which the VServer to create belongs to")
	flags.StringSliceVar(&vserverIDs, "vserver-id", nil, "Required. Resource ID of Vserver to update")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")

	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "vserver-id", func() []string {
		ulbID := ctx.PickResourceID(*req.ULBId)
		return getAllVServerIDNames(ctx, ulbID, *req.ProjectId, *req.Region)
	})

	return cmd
}

package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete returns ucloud ulb delete.
func newDelete(ctx *cli.Context) *cobra.Command {
	idNames := []string{}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDeleteULBRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete ULB instances by resource ID",
		Long:  "Delete ULB instances by resource ID",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.ULBId = sdk.String(id)
				_, err := client.DeleteULB(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "ulb[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "ulb-id", nil, "Required. Resource ID of the ULB instances to delete")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")

	return cmd
}

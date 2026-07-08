package vpc

import (
	"fmt"

	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete returns ucloud vpc delete.
func newDelete(ctx *cli.Context) *cobra.Command {
	idNames := []string{}
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDeleteVPCRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete vpc network",
		Long:    "Delete vpc network",
		Example: "ucloud vpc delete --vpc-id uvnet-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.VPCId = sdk.String(id)
				_, err := client.DeleteVPC(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "vpc[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().StringSliceVar(&idNames, "vpc-id", nil, "Required. Resource ID of the vpc network to delete")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("vpc-id")

	return cmd
}

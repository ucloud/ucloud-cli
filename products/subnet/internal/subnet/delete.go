package subnet

import (
	"fmt"

	"github.com/spf13/cobra"

	vpcsdk "github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete returns ucloud subnet delete.
func newDelete(ctx *cli.Context) *cobra.Command {
	idNames := []string{}
	client := cli.NewServiceClient(ctx, vpcsdk.NewClient)
	req := client.NewDeleteSubnetRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete subnet",
		Long:  "Delete subnet",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, id := range idNames {
				resourceID := ctx.PickResourceID(id)
				req.SubnetId = sdk.String(resourceID)
				_, err := client.DeleteSubnet(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "subnet[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: resourceID, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "subnet-id", nil, "Required. Resource ID of subent")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("subnet-id")
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, "", *req.ProjectId, *req.Region)
	})

	return cmd
}

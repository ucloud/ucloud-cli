package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBackendDelete returns ucloud ulb vserver backend delete.
func newBackendDelete(ctx *cli.Context) *cobra.Command {
	backendIDs := []string{}
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewReleaseBackendRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete ULB VServer backend nodes",
		Long:  "Delete ULB VServer backend nodes",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			results := []cli.OpResultRow{}
			for _, idname := range backendIDs {
				id := ctx.PickResourceID(idname)
				req.BackendId = sdk.String(id)
				_, err := client.ReleaseBackend(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "backend node[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete-backend", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB which the backend nodes belong to")
	flags.StringSliceVar(&backendIDs, "backend-id", nil, "Required. BackendID of backend nodes to update")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("backend-id")

	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "backend-id", func() []string {
		return getAllBackendNodeIDNames(ctx, *req.ULBId, "", *req.ProjectId, *req.Region)
	})
	return cmd
}

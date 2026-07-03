package udisk

import (
	"fmt"

	"github.com/spf13/cobra"

	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSnapshotDelete ucloud udisk delete-snapshot
func newSnapshotDelete(ctx *cli.Context) *cobra.Command {
	var snapshotIds *[]string
	client := cli.NewServiceClient(ctx, puhost.NewClient)
	req := client.NewDeleteSnapshotRequest()
	cmd := &cobra.Command{
		Use:   "delete-snapshot",
		Short: "Delete snapshots",
		Long:  "Delete snapshots",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, snapshotID := range *snapshotIds {
				req.SnapshotId = sdk.String(ctx.PickResourceID(snapshotID))
				resp, err := client.DeleteSnapshot(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(w, "snapshot[%s] deleted\n", resp.SnapshotId)
				results = append(results, cli.OpResultRow{ResourceID: resp.SnapshotId, Action: "delete-snapshot", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	snapshotIds = flags.StringSlice("snapshot-id", nil, "Required. Resource ID of snapshots to delete")
	flags.StringSliceVar(snapshotIds, "snaphost-id", nil, "Deprecated alias for --snapshot-id")
	flags.MarkHidden("snaphost-id")
	cmd.MarkFlagsOneRequired("snapshot-id", "snaphost-id")
	return cmd
}

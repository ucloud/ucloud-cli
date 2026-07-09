package usnap

import (
	"fmt"

	"github.com/spf13/cobra"

	usnapsdk "github.com/ucloud/ucloud-sdk-go/services/usnap"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDelete ucloud usnap delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes *bool
	client := cli.NewServiceClient(ctx, usnapsdk.NewClient)
	req := client.NewDeleteSnapshotServiceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete USnap snapshot service(s)",
		Long:  "Delete USnap snapshot service(s)",
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(*yes, "Are you sure to delete USnap snapshot service(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}

			_, err = client.DeleteSnapshotService(req)
			if err != nil {
				ctx.HandleError(err)
			} else {
				fmt.Fprintf(w, "usnap[%s] deleted\n", *req.VDiskId)
				results = append(results, cli.OpResultRow{ResourceID: *req.VDiskId, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.VDiskId = flags.String("vdisk-id", "", "Required. Resource ID of the disk whose snapshot service to delete")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	cmd.MarkFlagRequired("vdisk-id")

	return cmd
}

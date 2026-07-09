package upfs

import (
	"fmt"

	"github.com/spf13/cobra"

	upfssdk "github.com/ucloud/ucloud-sdk-go/services/upfs"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDelete ucloud upfs delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var volumeIDs *[]string
	client := cli.NewServiceClient(ctx, upfssdk.NewClient)
	req := client.NewRemoveUPFSVolumeRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete UPFS volume(s)",
		Long:  "Delete UPFS volume(s)",
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(*yes, "Are you sure to delete UPFS volume(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *volumeIDs {
				id := ctx.PickResourceID(id)
				req.VolumeId = &id
				_, err := client.RemoveUPFSVolume(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				} else {
					fmt.Fprintf(w, "upfs[%s] deleted\n", *req.VolumeId)
					results = append(results, cli.OpResultRow{ResourceID: *req.VolumeId, Action: "delete", Status: "Deleted"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	volumeIDs = flags.StringSlice("volume-id", nil, "Required. The Resource ID of UPFS volumes to delete")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	cmd.MarkFlagRequired("volume-id")

	return cmd
}

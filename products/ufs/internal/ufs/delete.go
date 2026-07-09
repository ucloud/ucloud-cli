package ufs

import (
	"fmt"

	"github.com/spf13/cobra"

	ufssdk "github.com/ucloud/ucloud-sdk-go/services/ufs"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDelete ucloud ufs delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var volumeIDs *[]string
	client := cli.NewServiceClient(ctx, ufssdk.NewClient)
	req := client.NewRemoveUFSVolumeRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete UFS volume(s)",
		Long:  "Delete UFS volume(s)",
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(*yes, "Are you sure to delete UFS volume(s)?")
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
				_, err := client.RemoveUFSVolume(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				} else {
					fmt.Fprintf(w, "ufs[%s] deleted\n", *req.VolumeId)
					results = append(results, cli.OpResultRow{ResourceID: *req.VolumeId, Action: "delete", Status: "Deleted"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	volumeIDs = flags.StringSlice("volume-id", nil, "Required. The Resource ID of UFS volumes to delete")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	ctx.BindCommonParams(cmd, req)

	cmd.MarkFlagRequired("volume-id")

	return cmd
}

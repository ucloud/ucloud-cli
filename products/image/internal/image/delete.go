package image

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud image delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var imageIDs *[]string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewTerminateCustomImageRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete custom images",
		Long:  "Delete custom images",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *imageIDs {
				req.ImageId = sdk.String(ctx.PickResourceID(id))
				resp, err := client.TerminateCustomImage(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(w, "image[%s] deleted\n", resp.ImageId)
				results = append(results, cli.OpResultRow{ResourceID: resp.ImageId, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	imageIDs = cmd.Flags().StringSlice("image-id", nil, "Required. Resource ID of images")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	cmd.MarkFlagRequired("image-id")
	command.SetCompletion(cmd, "image-id", func() []string {
		return getImageList(ctx, []string{IMAGE_AVAILABLE, IMAGE_COPYING, IMAGE_MAKING}, IAMGE_CUSTOM, *req.ProjectId, *req.Region, "")
	})
	return cmd
}

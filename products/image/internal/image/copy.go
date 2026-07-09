package image

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCopy ucloud image copy
func newCopy(ctx *cli.Context) *cobra.Command {
	var imageIDs *[]string
	var async *bool
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewCopyCustomImageRequest()
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy custom images",
		Long:  "Copy custom images",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			*req.ProjectId = ctx.PickResourceID(*req.ProjectId)
			*req.TargetProjectId = ctx.PickResourceID(*req.TargetProjectId)
			for _, id := range *imageIDs {
				id = ctx.PickResourceID(id)
				req.SourceImageId = &id
				resp, err := client.CopyCustomImage(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				text := fmt.Sprintf("image[%s] is coping", resp.TargetImageId)
				if *async {
					fmt.Fprintln(w, text)
				} else {
					// M2: poll the TARGET project/region (not the source request
					// defaults) so cross-region copy converges. Mirrors cmd/image.go,
					// whose base.Poller.Poll bound *req.TargetProjectId/*req.TargetRegion.
					ctx.PollerTo(w, describeImageByID(ctx, *req.TargetProjectId, *req.TargetRegion, "")).Spoll(resp.TargetImageId, text, []string{IMAGE_AVAILABLE, IMAGE_UNAVAILABLE})
				}
				results = append(results, cli.OpResultRow{ResourceID: resp.TargetImageId, Action: "copy", Status: "Copying"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	imageIDs = cmd.Flags().StringSlice("source-image-id", nil, "Required. Resource ID of source image")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.TargetRegion = flags.String("target-region", ctx.DefaultRegion(), "Optional. Target region. See 'ucloud region'")
	req.TargetProjectId = flags.String("target-project", ctx.DefaultProjectID(), "Optional. Target Project ID. See 'ucloud project list'")
	req.TargetImageName = flags.String("target-image-name", "", "Optional. Name of target image")
	req.TargetImageDescription = flags.String("target-image-desc", "", "Optional. Description of target image")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")

	command.SetCompletion(cmd, "source-image-id", func() []string {
		return getImageList(ctx, []string{IMAGE_AVAILABLE}, IAMGE_CUSTOM, *req.ProjectId, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "zone", func() []string {
		return ctx.ZoneList(*req.Region)
	})
	command.SetCompletion(cmd, "target-region", ctx.RegionList)
	command.SetCompletion(cmd, "target-project", ctx.ProjectList)

	cmd.MarkFlagRequired("source-image-id")

	return cmd
}

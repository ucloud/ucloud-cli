package uhost

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreateImage ucloud uhost create-image
func newCreateImage(ctx *cli.Context) *cobra.Command {
	var async *bool
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewCreateCustomImageRequest()
	cmd := &cobra.Command{
		Use:   "create-image",
		Short: "Create image from an uhost instance",
		Long:  "Create image from an uhost instance",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			req.UHostId = sdk.String(ctx.PickResourceID(*req.UHostId))
			resp, err := client.CreateCustomImage(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			// "iamge[%s] is making" typo preserved verbatim from cmd/uhost.go.
			text := fmt.Sprintf("iamge[%s] is making", resp.ImageId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeImageByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.ImageId, text, []string{IMAGE_AVAILABLE, IMAGE_UNAVAILABLE})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.ImageId, Action: "create", Status: "Making"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.UHostId = flags.String("uhost-id", "", "Resource ID of uhost to create image from")
	req.ImageName = flags.String("image-name", "", "Required. Name of the image to create")
	req.ImageDescription = flags.String("image-desc", "", "Optional. Description of the image to create")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("image-name")
	return cmd
}

package image

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	cliconst "github.com/ucloud/ucloud-cli/model/cli"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// NewCommand builds the `image` root command and mounts the 4 subcommands.
// Mirrors cmd/image.go NewCmdUImage (same AddCommand order: list, copy, delete,
// create). The create subcommand is image's OWN copy of uhost's create-image
// (newCreateImage) — image no longer borrows NewCmdUhostCreateImage from cmd.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "List and manipulate images",
		Long:  `List and manipulate images`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCopy(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newCreateImage(ctx))

	return cmd
}

// newList ucloud image list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeImageRequest()
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List image",
		Long:    "List image",
		Example: "ucloud image list --image-type Base",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeImage(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := make([]ImageRow, 0)
			for _, image := range resp.ImageSet {
				row := ImageRow{}
				row.ImageName = image.ImageName
				row.ImageID = image.ImageId
				row.ImageType = image.ImageType
				row.BasicImage = image.OsName
				row.ExtensibleFeature = strings.Join(image.Features, ",")
				row.CreationTime = common.FormatDate(image.CreateTime)
				row.State = image.State
				if row.State == "Available" {
					list = append(list, row)
				}
			}
			ctx.PrintList(list)
		},
	}
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.ImageType = cmd.Flags().String("image-type", "Base", "Optional. 'Base',Standard image; 'Business',image market; 'Custom',custom image")
	req.OsType = cmd.Flags().String("os-type", "", "Optional. Linux or Windows. Return all types by default")
	req.ImageId = cmd.Flags().String("image-id", "", "Optional. Resource ID of image")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 500, "Optional. Max count")
	command.SetFlagValues(cmd, "image-type", "Base", "Business", "Custom")
	return cmd
}

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
		return getImageList(ctx, []string{status.IMAGE_AVAILABLE, status.IMAGE_COPYING, status.IMAGE_MAKING}, cliconst.IAMGE_CUSTOM, *req.ProjectId, *req.Region, "")
	})
	return cmd
}

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
					ctx.PollerTo(w, describeImageByID(ctx, *req.TargetProjectId, *req.TargetRegion, "")).Spoll(resp.TargetImageId, text, []string{status.IMAGE_AVAILABLE, status.IMAGE_UNAVAILABLE})
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
		return getImageList(ctx, []string{status.IMAGE_AVAILABLE}, cliconst.IAMGE_CUSTOM, *req.ProjectId, *req.Region, *req.Zone)
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

// newCreateImage ucloud image create — image's OWN copy of uhost's create-image.
// Copied verbatim from cmd/uhost.go NewCmdUhostCreateImage (CreateCustomImage +
// poll); Use is "create" to match the original `image create` (which borrowed
// uhost's command and renamed its Use). uhost (Part 6) keeps its own
// create-image — duplication across products is allowed (product autonomy).
func newCreateImage(ctx *cli.Context) *cobra.Command {
	var async *bool
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewCreateCustomImageRequest()
	cmd := &cobra.Command{
		Use:   "create",
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
			// M2: typo "iamge[%s] is making" preserved verbatim from cmd/uhost.go.
			text := fmt.Sprintf("iamge[%s] is making", resp.ImageId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeImageByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.ImageId, text, []string{status.IMAGE_AVAILABLE, status.IMAGE_UNAVAILABLE})
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
		return getUhostList(ctx, []string{status.HOST_RUNNING, status.HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("uhost-id")
	cmd.MarkFlagRequired("image-name")
	return cmd
}

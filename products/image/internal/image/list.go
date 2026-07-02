package image

import (
	"strings"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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

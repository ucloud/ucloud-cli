package css

import (
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDiskLimit ucloud css disk-limit
func newDiskLimit(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewGetUESDiskSizeLimitationRequest()
	cmd := &cobra.Command{
		Use:   "disk-limit",
		Short: "List UES disk size limitations by disk type",
		Long:  "List UES disk size limitations by disk type",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.GetUESDiskSizeLimitation(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []DiskLimitRow{}
			for _, d := range resp.DiskSizeLimitationSet {
				list = append(list, DiskLimitRow{
					DiskType:  d.DiskType,
					MinSizeGB: fmt.Sprintf("%d", d.MinSize),
					MaxSizeGB: fmt.Sprintf("%d", d.MaxSize),
				})
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	return cmd
}

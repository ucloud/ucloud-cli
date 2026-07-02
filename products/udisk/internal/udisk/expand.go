package udisk

import (
	"fmt"

	"github.com/spf13/cobra"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newExpand ucloud udisk expand
func newExpand(ctx *cli.Context) *cobra.Command {
	var udiskIDs *[]string
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewResizeUDiskRequest()
	cmd := &cobra.Command{
		Use:   "expand",
		Short: "Expand udisk size",
		Long:  "Expand udisk size",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *udiskIDs {
				id = ctx.PickResourceID(id)
				req.UDiskId = &id
				_, err := client.ResizeUDisk(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(w, "udisk:[%s] expanded to %d GB\n", *req.UDiskId, *req.Size)
				results = append(results, cli.OpResultRow{ResourceID: *req.UDiskId, Action: "expand", Status: "Expanded"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of the udisks to expand")
	req.Size = flags.Int("size-gb", 0, "Required. Size of the udisk after expanded. Unit: GB. Range [1,8000]")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")

	command.SetCompletion(cmd, "udisk-id", func() []string {
		return getDiskList(ctx, []string{DISK_AVAILABLE}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")
	cmd.MarkFlagRequired("size-gb")

	return cmd
}

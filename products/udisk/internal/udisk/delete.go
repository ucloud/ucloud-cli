package udisk

import (
	"fmt"

	"github.com/spf13/cobra"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud udisk delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var udiskIDs *[]string
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewDeleteUDiskRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete udisk instances",
		Long:  "Delete udisk instances",
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(*yes, "Are you sure to delete udisk(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *udiskIDs {
				id := ctx.PickResourceID(id)
				req.UDiskId = &id
				_, err := client.DeleteUDisk(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				} else {
					fmt.Fprintf(w, "udisk[%s] deleted\n", *req.UDiskId)
					results = append(results, cli.OpResultRow{ResourceID: *req.UDiskId, Action: "delete", Status: "Deleted"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. The Resource ID of udisks to delete")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")

	command.SetCompletion(cmd, "udisk-id", func() []string {
		return getDiskList(ctx, []string{DISK_AVAILABLE, DISK_FAILED}, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udisk-id")

	return cmd
}

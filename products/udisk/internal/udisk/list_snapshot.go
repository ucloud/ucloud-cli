package udisk

import (
	"fmt"

	"github.com/spf13/cobra"

	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSnapshotList ucloud udisk list-snapshot
func newSnapshotList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, puhost.NewClient)
	req := client.NewDescribeSnapshotRequest()
	cmd := &cobra.Command{
		Use:   "list-snapshot",
		Short: "List snapshots",
		Long:  "List snapshots",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeSnapshot(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []SnapshotRow{}
			for _, snapshot := range resp.UHostSnapshotSet {
				row := SnapshotRow{
					Name:             snapshot.SnapshotName,
					ResourceID:       snapshot.SnapshotId,
					AvailabilityZone: snapshot.Zone,
					BoundUDisk:       snapshot.DiskId,
					Size:             fmt.Sprintf("%dGB", snapshot.Size),
					State:            snapshot.State,
					UDiskType:        snapshot.DiskType,
					CreationTime:     common.FormatDate(snapshot.CreateTime),
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	// StringSliceVar binds the flag to req.SnapshotIds so Cobra fills it during
	// parse; dereferencing StringSlice() here would freeze it to the initial nil
	// slice and drop the --snapshot-id filter.
	flags.StringSliceVar(&req.SnapshotIds, "snapshot-id", nil, "Optional. Resource ID of snapshots to list")
	req.UHostId = flags.String("uhost-id", "", "Optional. Snapshots of the uhost")
	req.DiskId = flags.String("disk-id", "", "Optional. Snapshots of the udisk")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit, length of snapshot list")

	return cmd
}

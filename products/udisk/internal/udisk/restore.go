package udisk

import (
	"fmt"

	"github.com/spf13/cobra"

	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRestore ucloud udisk restore
func newRestore(ctx *cli.Context) *cobra.Command {
	var snapshotIDs *[]string
	var yes *bool
	client := cli.NewServiceClient(ctx, puhost.NewClient)
	req := client.NewRestoreUHostDiskRequest()
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore udisk from snapshot",
		Long:  "Restore udisk from snapshot",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, snapshotID := range *snapshotIDs {
				snapshotID = ctx.PickResourceID(snapshotID)
				any, err := describeSnapshotByID(ctx)(snapshotID, nil)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				snapshot, ok := any.(*puhost.SnapshotSet)
				if !ok {
					fmt.Fprintf(w, "snapshot[%s] doesn't exist\n", snapshotID)
					continue
				}
				if snapshot.UHostId != "" {
					text := fmt.Sprintf("can we detach udisk[%s] from uhost[%s]?", snapshot.DiskId, snapshot.UHostId)
					ok, err := ctx.Confirm(*yes, text)
					if err != nil {
						ctx.HandleError(err)
						continue
					}
					if !ok {
						continue
					}
					DetachUdisk(ctx, false, snapshot.DiskId, w)
				}
				req.SnapshotIds = append(req.SnapshotIds, snapshotID)
				_, err = client.RestoreUHostDisk(req)

				if err != nil {
					ctx.HandleError(err)
					return
				}

				text := fmt.Sprintf("udisk[%s] has been restored from snapshot[%s]", snapshot.DiskId, snapshot.SnapshotId)
				fmt.Fprintln(w, text)
				results = append(results, cli.OpResultRow{ResourceID: snapshot.DiskId, Action: "restore", Status: "Restored"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	snapshotIDs = flags.StringSlice("snapshot-id", nil, "Required. Resourece ID of the snapshots to restore from")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	command.SetCompletion(cmd, "snapshot-id", func() []string {
		return getSnapshotList(ctx, []string{SNAPSHOT_NORMAL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("snapshot-id")
	return cmd
}

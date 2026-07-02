package udisk

import (
	"fmt"

	"github.com/spf13/cobra"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newSnapshot ucloud udisk snapshot
func newSnapshot(ctx *cli.Context) *cobra.Command {
	var async *bool
	var udiskIDs *[]string
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewCreateUDiskSnapshotRequest()
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Create shapshots for udisks",
		Long:  "Create shapshots for udisks",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range *udiskIDs {
				id = ctx.PickResourceID(id)
				req.UDiskId = &id
				resp, err := client.CreateUDiskSnapshot(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				if len(resp.SnapshotId) == 1 {
					text := fmt.Sprintf("snapshot[%s] is creating", resp.SnapshotId[0])
					if *async {
						fmt.Fprintln(w, text)
					} else {
						ctx.PollerTo(w, describeSnapshotByID(ctx)).Spoll(resp.SnapshotId[0], text, []string{SNAPSHOT_NORMAL})
					}
					results = append(results, cli.OpResultRow{ResourceID: resp.SnapshotId[0], Action: "snapshot", Status: "Creating"})
				} else {
					fmt.Fprintf(w, "snapshot%v is creating. expect snapshot count 1, accept %d\n", resp.SnapshotId, len(resp.SnapshotId))
					for _, sid := range resp.SnapshotId {
						results = append(results, cli.OpResultRow{ResourceID: sid, Action: "snapshot", Status: "Creating"})
					}
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	udiskIDs = flags.StringSlice("udisk-id", nil, "Required. Resource ID of udisks to snapshot")
	req.Name = flags.String("name", "", "Required. Name of snapshots")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.Comment = flags.String("comment", "", "Optional. Description of snapshots")
	async = flags.BoolP("async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "udisk-id", func() []string {
		return getDiskList(ctx, []string{DISK_AVAILABLE, DISK_INUSE}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("udisk-id")
	cmd.MarkFlagRequired("name")
	return cmd
}

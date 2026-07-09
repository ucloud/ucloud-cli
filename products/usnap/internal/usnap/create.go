package usnap

import (
	"fmt"

	"github.com/spf13/cobra"

	usnapsdk "github.com/ucloud/ucloud-sdk-go/services/usnap"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud usnap create
func newCreate(ctx *cli.Context) *cobra.Command {
	var async *bool
	client := cli.NewServiceClient(ctx, usnapsdk.NewClient)
	req := client.NewCreateSnapshotServiceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a USnap snapshot service for a disk",
		Long:  "Create a USnap snapshot service for a disk",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			resp, err := client.CreateSnapshotService(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			text := fmt.Sprintf("usnap:%v is creating", resp.SnapshotServiceId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUsnapByID(ctx)).Spoll(resp.SnapshotServiceId, text, []string{SERVICE_AVAILABLE, SERVICE_FAILED})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.SnapshotServiceId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.VDiskId = flags.String("vdisk-id", "", "Required. Resource ID of the disk to create snapshot service for")
	req.BackupMode = flags.String("backup-mode", "", "Optional. Backup mode")
	req.Day = flags.Int("backup-day", 0, "Optional. Backup day range")
	req.Hour = flags.Int("backup-hour", 0, "Optional. Backup hour")
	req.Journal = flags.Int("journal", 0, "Optional. Journal retention count")
	req.ChargeType = flags.String("charge-type", "Dynamic", "Optional. 'Year', pay yearly; 'Month', pay monthly; 'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")

	cmd.MarkFlagRequired("vdisk-id")

	return cmd
}

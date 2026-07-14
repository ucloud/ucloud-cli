package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBackupUpdateStrategy ucloud pgsql backup update-strategy
func newBackupUpdateStrategy(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewUpdateUPgSQLBackupStrategyRequest()
	cmd := &cobra.Command{
		Use:   "update-strategy",
		Short: "Update the backup strategy of a UPgSQL instance",
		Long:  "Update the backup strategy of a UPgSQL instance",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			_, err := client.UpdateUPgSQLBackupStrategy(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "backup strategy of pgsql[%s] updated\n", *req.InstanceID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.InstanceID, Action: "update-strategy", Status: "Updated"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	req.BackupTimeRange = flags.String("backup-time-range", "", "Optional. Auto backup start time range, e.g. (3:00~4:00)")
	req.BackupWeek = flags.String("backup-week", "", "Optional. Days of week to start auto backup, e.g. 1,2,3,4,5,6,7")
	req.BackupMethod = flags.String("backup-method", "", "Optional. Default backup method")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("instance-id")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}

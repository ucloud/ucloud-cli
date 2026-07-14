package pgsql

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

var pgsqlLogTypeList = []string{"slow", "error"}

// newLogBackup ucloud pgsql log backup
func newLogBackup(ctx *cli.Context) *cobra.Command {
	var beginTime, endTime string
	client := newUPgSQLClient(ctx)
	req := client.NewBackupUPgSQLLogRequest()
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Back up the log package of a UPgSQL instance",
		Long:  "Back up the slow/error log package of a UPgSQL instance",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			if beginTime != "" {
				bt, err := time.Parse(common.DateTimeLayout, beginTime)
				if err != nil {
					ctx.HandleError(fmt.Errorf("invalid begin-time (use %s): %w", common.DateTimeLayout, err))
					return
				}
				req.BeginTime = sdk.Int(int(bt.Unix()))
			}
			if endTime != "" {
				et, err := time.Parse(common.DateTimeLayout, endTime)
				if err != nil {
					ctx.HandleError(fmt.Errorf("invalid end-time (use %s): %w", common.DateTimeLayout, err))
					return
				}
				req.EndTime = sdk.Int(int(et.Unix()))
			}
			_, err := client.BackupUPgSQLLog(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "log of pgsql[%s] backuped\n", *req.InstanceID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.InstanceID, Action: "log-backup", Status: "Backuped"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	req.BackupName = flags.String("name", "", "Required. Name of the exported backup file")
	req.BackupFile = flags.String("backup-file", "", "Required. Name of the log query result file")
	req.LogType = flags.String("log-type", "", "Optional. Log type. Accept values: slow, error")
	flags.StringVar(&beginTime, "begin-time", "", "Optional. Log begin time, e.g. 2019-01-02/15:04:05")
	flags.StringVar(&endTime, "end-time", "", "Optional. Log end time, e.g. 2019-01-02/15:04:05")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	command.SetFlagValues(cmd, "log-type", pgsqlLogTypeList...)
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("backup-file")

	return cmd
}

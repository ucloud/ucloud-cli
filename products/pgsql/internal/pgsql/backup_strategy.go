package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBackupStrategy ucloud pgsql backup strategy
func newBackupStrategy(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewGetUPgSQLBackupStrategyRequest()
	cmd := &cobra.Command{
		Use:   "strategy",
		Short: "Display the backup strategy of a UPgSQL instance",
		Long:  "Display the backup strategy of a UPgSQL instance",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			resp, err := client.GetUPgSQLBackupStrategy(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintln(ctx.ProgressWriter(), "BackupStrategy:")
			ctx.PrintList([]PgsqlBackupStrategyRow{{
				BackupMethod:    resp.BackupMethod,
				BackupTimeRange: resp.BackupTimeRange,
				BackupWeek:      resp.BackupWeek,
			}})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("instance-id")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}

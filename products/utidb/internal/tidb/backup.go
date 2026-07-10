package tidb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBackup ucloud utidb backup
func newBackup(ctx *cli.Context) *cobra.Command {
	var id, backupFilter, backupTs string

	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewStartTiDBClusterBackupRequest()

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Start a backup of a UTiDB instance",
		Long:  "Start a backup of a UTiDB instance",
		Run: func(c *cobra.Command, args []string) {
			req.Id = sdk.String(ctx.PickResourceID(id))
			if backupFilter != "" {
				req.BackupFilter = sdk.String(backupFilter)
			}
			if backupTs != "" {
				req.BackupTs = sdk.String(backupTs)
			}
			resp, err := client.StartTiDBClusterBackup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.BackupId, Action: "backup", Status: stateBackingUp})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&id, "utidb-id", "", "Required. Resource ID of the UTiDB instance")
	flags.StringVar(&backupFilter, "backup-filter", "", "Optional. Backup filter rule")
	flags.StringVar(&backupTs, "backup-ts", "", "Optional. Backup timestamp")

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("utidb-id")
	command.SetCompletion(cmd, "utidb-id", func() []string {
		return listResourceIDs(ctx, nil, req.GetRegion(), req.GetZone(), req.GetProjectId())
	})

	return cmd
}

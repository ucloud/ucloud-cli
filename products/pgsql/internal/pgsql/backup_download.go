package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newBackupDownload ucloud pgsql backup download
func newBackupDownload(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewGetUPgSQLBackupURLRequest()
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Display download URLs of a UPgSQL backup",
		Long:  "Display the public and inner download URLs of a UPgSQL backup",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			resp, err := client.GetUPgSQLBackupURL(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.PrintList([]PgsqlBackupURLRow{{
				BackupPath:      resp.BackupPath,
				InnerBackupPath: resp.InnerBackupPath,
			}})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	req.BackupID = flags.String("backup-id", "", "Required. Backup ID of the backup to download")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("backup-id")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})
	command.SetCompletion(cmd, "backup-id", func() []string {
		return getBackupIDList(ctx, *req.InstanceID, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}

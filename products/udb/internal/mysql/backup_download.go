package mysql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBBackupGetDownloadURL ucloud udb backup get-download-url
func newUDBBackupGetDownloadURL(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBInstanceBackupURLRequest()
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Display download url of backup",
		Long:  "Display download url of backup",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeUDBInstanceBackupURL(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintln(ctx.Out(), resp.BackupPath)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.BackupId = flags.Int("backup-id", -1, "Required. BackupID of backup to delete")
	req.DBId = flags.String("udb-id", "", "Required. Resource ID of udb which the backup belongs to")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("udb-id")
	cmd.MarkFlagRequired("backup-id")
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

package mysql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBBackupCreate ucloud udb backup create
func newUDBBackupCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewBackupUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create backups for MySQL instance manually",
		Long:  "Create backups for MySQL instance manually",
		Run: func(c *cobra.Command, args []string) {
			*req.DBId = ctx.PickResourceID(*req.DBId)
			_, err := client.BackupUDBInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "udb[%s] backuped\n", *req.DBId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.DBId, Action: "create", Status: "Backuped"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBId = flags.String("udb-id", "", "Required. Resource ID of UDB instnace to backup")
	req.BackupName = flags.String("name", "", "Required. Name of backup")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("udb-id")
	cmd.MarkFlagRequired("name")

	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

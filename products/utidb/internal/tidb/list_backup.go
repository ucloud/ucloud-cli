package tidb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListBackup ucloud utidb list-backup
func newListBackup(ctx *cli.Context) *cobra.Command {
	var id string
	var limit, offset int

	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewListTiDBClusterBackupRequest()

	cmd := &cobra.Command{
		Use:   "list-backup",
		Short: "List backups of a UTiDB instance",
		Long:  "List backups of a UTiDB instance",
		Run: func(c *cobra.Command, args []string) {
			req.Id = sdk.String(ctx.PickResourceID(id))
			req.Limit = sdk.Int(limit)
			req.Offset = sdk.Int(offset)
			resp, err := client.ListTiDBClusterBackup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []backupRow{newBackupRowFromData(resp.Data)}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&id, "utidb-id", "", "Required. Resource ID of the UTiDB instance")
	flags.IntVar(&limit, "limit", 30, "Optional. The maximum number of resources per page")
	flags.IntVar(&offset, "offset", 0, "Optional. The index of resource which start to list")

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("utidb-id")
	command.SetCompletion(cmd, "utidb-id", func() []string {
		return listResourceIDs(ctx, nil, req.GetRegion(), req.GetZone(), req.GetProjectId())
	})

	return cmd
}

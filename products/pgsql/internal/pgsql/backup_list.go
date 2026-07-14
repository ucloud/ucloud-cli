package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

var pgsqlBackupTypeMap = map[string]int{
	"auto":   1,
	"manual": 2,
}
var pgsqlReverseBackupTypeMap = map[int]string{
	1: "auto",
	2: "manual",
	0: "all",
}

// newBackupList ucloud pgsql backup list
func newBackupList(ctx *cli.Context) *cobra.Command {
	var bpType string
	client := newUPgSQLClient(ctx)
	req := client.NewListUPgSQLBackupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backups of a UPgSQL instance",
		Long:  "List backups of a UPgSQL instance",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			if v, ok := pgsqlBackupTypeMap[bpType]; ok {
				req.BackupType = sdk.Int(v)
			}
			resp, err := client.ListUPgSQLBackup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []PgsqlBackupRow{}
			for _, b := range resp.DataSet {
				list = append(list, PgsqlBackupRow{
					BackupID:        b.BackupID,
					BackupName:      b.BackupName,
					InstanceID:      b.InstanceID,
					State:           b.State,
					BackupType:      pgsqlReverseBackupTypeMap[b.BackupType],
					BackupSize:      fmt.Sprintf("%dB", b.BackupSize),
					BackupStartTime: common.FormatDateTime(b.BackupStartTime),
					BackupEndTime:   common.FormatDateTime(b.BackupEndTime),
				})
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	flags.StringVar(&bpType, "backup-type", "", "Optional. Backup type. Accept values: auto, manual")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindOffset(cmd, req)
	ctx.BindLimit(cmd, req)

	command.SetFlagValues(cmd, "backup-type", "auto", "manual")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	cmd.MarkFlagRequired("instance-id")

	return cmd
}

package mysql

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type udbBackupRow struct {
	BackupID         int
	BackupName       string
	DB               string
	BackupSize       string
	BackupType       string
	Status           string
	AvailabilityZone string
	BackupBeginTime  string
	BackupEndTime    string
}

var dbTypeMap = map[string]string{
	"mysql":      "sql",
	"mongodb":    "nosql",
	"postgresql": "postgresql",
	"sqlserver":  "sqlserver",
}

var dbTypeList = []string{"mysql", "mongodb", "postgresql", "sqlserver"}

// newUDBBackupList ucloud udb backup list
func newUDBBackupList(ctx *cli.Context) *cobra.Command {
	var bpType, dbType, beginTime, endTime, backupID string
	bpTypeMap := map[string]int{
		"manual": 1,
		"auto":   0,
	}
	reverseBpTypeMap := map[int]string{
		1: "manual",
		0: "auto",
	}
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBBackupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backups of MySQL instance",
		Long:  "List backups of MySQL instance",
		Run: func(c *cobra.Command, args []string) {
			if v, ok := bpTypeMap[bpType]; ok {
				req.BackupType = &v
			}
			if v, ok := dbTypeMap[dbType]; ok {
				req.ClassType = &v
			}
			if *req.DBId != "" {
				*req.DBId = ctx.PickResourceID(*req.DBId)
			}
			if backupID != "" {
				id, err := strconv.Atoi(ctx.PickResourceID(backupID))
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.BackupId = &id
			}
			if beginTime != "" {
				bt, err := time.Parse("2006-01-02/15:04:05", beginTime)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.BeginTime = sdk.Int(int(bt.Unix()))
			}
			if endTime != "" {
				bt, err := time.Parse("2006-01-02/15:04:05", endTime)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.EndTime = sdk.Int(int(bt.Unix()))
			}
			resp, err := client.DescribeUDBBackup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []udbBackupRow{}
			for _, ins := range resp.DataSet {
				row := udbBackupRow{
					BackupID:         ins.BackupId,
					BackupName:       ins.BackupName,
					AvailabilityZone: ins.Zone,
					DB:               fmt.Sprintf("%s|%s", ins.DBName, ins.DBId),
					BackupSize:       fmt.Sprintf("%dB", ins.BackupSize),
					BackupType:       reverseBpTypeMap[ins.BackupType],
					Status:           ins.State,
					BackupBeginTime:  common.FormatDateTime(ins.BackupTime),
					BackupEndTime:    common.FormatDateTime(ins.BackupEndTime),
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBId = flags.String("udb-id", "", "Optional. Resource ID of UDB for list the backups of the specifid UDB")
	flags.StringVar(&backupID, "backup-id", "", "Optional. Resource ID of backup. List the specified backup only")
	flags.StringVar(&bpType, "backup-type", "", "Optional. Backup type. Accept values:auto or manual")
	flags.StringVar(&dbType, "db-type", "", "Optional. Only list backups of the UDB of the specified DB type")
	flags.StringVar(&beginTime, "begin-time", "", "Optional. Begin time of backup. For example, 2019-02-26/11:21:39")
	flags.StringVar(&endTime, "end-time", "", "Optional. End time of backup. For example, 2019-02-26/11:31:39")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindOffset(cmd, req)
	ctx.BindLimit(cmd, req)

	command.SetFlagValues(cmd, "backup-type", "auto", "manual")
	command.SetFlagValues(cmd, "db-type", dbTypeList...)
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

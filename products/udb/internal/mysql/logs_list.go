package mysql

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type udbArchiveRow struct {
	ArchiveID  int
	Name       string
	LogType    string
	DB         string
	Size       string
	Status     string
	CreateTime string
}

// newUDBLogArchiveList ucloud udb log archive list
func newUDBLogArchiveList(ctx *cli.Context) *cobra.Command {
	var beginTime, endTime string
	logTypes := []string{}
	logTypeMap := map[string]int{
		"binlog":     2,
		"slow_query": 3,
		"error":      4,
	}
	rLogTypeMap := map[int]string{
		2: "binlog",
		3: "slow_query",
		4: "error",
	}
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBLogPackageRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mysql log archives(log files)",
		Long:  "List mysql log archives(log files)",
		Run: func(c *cobra.Command, args []string) {
			if beginTime != "" {
				bt, err := time.Parse(common.DateTimeLayout, beginTime)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.BeginTime = sdk.Int(int(bt.Unix()))
			}
			if endTime != "" {
				et, err := time.Parse(common.DateTimeLayout, endTime)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.EndTime = sdk.Int(int(et.Unix()))
			}

			if *req.DBId != "" {
				*req.DBId = ctx.PickResourceID(*req.DBId)
			}

			for _, s := range logTypes {
				if v, ok := logTypeMap[s]; ok {
					req.Types = append(req.Types, v)
				} else {
					ctx.HandleError(fmt.Errorf("log-type should be one of 'binlog', 'slow_query' or 'error'"))
				}
			}

			resp, err := client.DescribeUDBLogPackage(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []udbArchiveRow{}
			for _, ins := range resp.DataSet {
				row := udbArchiveRow{
					ArchiveID:  ins.BackupId,
					Name:       ins.BackupName,
					LogType:    rLogTypeMap[ins.BackupType],
					DB:         fmt.Sprintf("%s|%s", ins.DBId, ins.DBName),
					Size:       fmt.Sprintf("%dB", ins.BackupSize),
					Status:     ins.State,
					CreateTime: common.FormatDateTime(ins.BackupTime),
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&logTypes, "log-type", nil, "Optional. Type of log. Accept Values: binlog, slow_query and error")
	req.DBId = flags.String("udb-id", "", "Optional. Resource ID of UDB instance which the listed logs belong to")
	flags.StringVar(&beginTime, "begin-time", "", "Optional. For example 2019-01-02/15:04:05")
	flags.StringVar(&endTime, "end-time", "", "Optional. For example 2019-01-02/15:04:05")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindLimit(cmd, req)
	ctx.BindOffset(cmd, req)

	command.SetFlagValues(cmd, "log-type", "binlog", "slow_query", "error")
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

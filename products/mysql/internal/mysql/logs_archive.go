package mysql

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBLogArchiveCreate ucloud udb log archive create
func newUDBLogArchiveCreate(ctx *cli.Context) *cobra.Command {
	var udbID string
	var name, logType, beginTime, endTime string
	var commonBase request.CommonBase
	cmd := &cobra.Command{
		Use:     "archive",
		Short:   "Archive the log of mysql as a compressed file",
		Long:    "Archive the log of mysql as a compressed file",
		Example: "ucloud mysql logs archive --name test.cli2 --udb-id udb-xxx/test.cli1 --log-type slow_query --begin-time 2019-02-23/15:30:00 --end-time 2019-02-24/15:31:00",
		Run: func(c *cobra.Command, args []string) {
			region := commonBase.GetRegion()
			zone := commonBase.GetZone()
			project := commonBase.GetProjectId()
			udbID = ctx.PickResourceID(udbID)
			client := cli.NewServiceClient(ctx, udb.NewClient)
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			if logType == "slow_query" {
				if beginTime == "" || endTime == "" {
					ctx.HandleError(fmt.Errorf("both begin-time and end-time can not be empty"))
					return
				}
				bt, err := time.Parse(common.DateTimeLayout, beginTime)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				et, err := time.Parse(common.DateTimeLayout, endTime)
				if err != nil {
					ctx.HandleError(err)
					return
				}

				req := client.NewBackupUDBInstanceSlowLogRequest()
				req.BeginTime = sdk.Int(int(bt.Unix()))
				req.EndTime = sdk.Int(int(et.Unix()))
				req.DBId = &udbID
				req.BackupName = &name
				req.Region = &region
				req.ProjectId = &project

				_, err = client.BackupUDBInstanceSlowLog(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(w, "mysql log archive[%s] created\n", name)
				results = append(results, cli.OpResultRow{ResourceID: name, Action: "archive", Status: "Created"})
			} else if logType == "error" {
				req := client.NewBackupUDBInstanceErrorLogRequest()
				req.DBId = &udbID
				req.BackupName = &name
				req.Region = &region
				req.Zone = &zone
				req.ProjectId = &project

				_, err := client.BackupUDBInstanceErrorLog(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(w, "mysql log archive[%s] created\n", name)
				results = append(results, cli.OpResultRow{ResourceID: name, Action: "archive", Status: "Created"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&udbID, "udb-id", "", "Required. Resource ID of UDB instance which we fetch logs from")
	flags.StringVar(&name, "name", "", "Required. Name of compressed file")
	flags.StringVar(&logType, "log-type", "", "Required. Type of log to package. Accept values: slow_query, error")
	flags.StringVar(&beginTime, "begin-time", "", "Optional. Required when log-type is slow. For example 2019-01-02/15:04:05")
	flags.StringVar(&endTime, "end-time", "", "Optional. Required when log-type is slow. For example 2019-01-02/15:04:05")
	ctx.BindRegion(cmd, &commonBase)
	ctx.BindZone(cmd, &commonBase)
	ctx.BindProjectID(cmd, &commonBase)

	cmd.MarkFlagRequired("udb-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("log-type")

	command.SetFlagValues(cmd, "log-type", "slow_query", "error")
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sql", commonBase.GetProjectId(), commonBase.GetRegion(), commonBase.GetZone())
	})
	return cmd
}

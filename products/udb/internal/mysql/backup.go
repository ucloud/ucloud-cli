package mysql

import (
	"fmt"
	"io/ioutil" //nolint:staticcheck // keep ioutil for zero-behavior-change verbatim copy
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// dateTimeLayout 时间格式，mirrors base.DateTimeLayout.
const dateTimeLayout = "2006-01-02/15:04:05"

// formatDateTime mirrors base.FormatDateTime: format a unix second timestamp.
func formatDateTime(seconds int) string {
	return time.Unix(int64(seconds), 0).Format("2006-01-02/15:04:05")
}

// getHomePath mirrors base.GetHomePath.
func getHomePath() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// getFileList mirrors base.GetFileList: complete file names for shell completion.
func getFileList(suffix string) []string {
	cmdLine := strings.TrimSpace(os.Getenv("COMP_LINE"))
	words := strings.Split(cmdLine, " ")
	last := words[len(words)-1]
	pathPrefix := "."

	if !strings.HasPrefix(last, "-") {
		pathPrefix = last
	}
	hasTilde := false
	//https://tiswww.case.edu/php/chet/bash/bashref.html#Tilde-Expansion
	if strings.HasPrefix(pathPrefix, "~") {
		pathPrefix = strings.Replace(pathPrefix, "~", getHomePath(), 1)
		hasTilde = true
	}
	files, err := ioutil.ReadDir(pathPrefix)
	if err != nil {
		return nil
	}
	names := []string{}
	for _, f := range files {
		name := f.Name()
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		if hasTilde {
			pathPrefix = strings.Replace(pathPrefix, getHomePath(), "~", 1)
		}
		if strings.HasSuffix(pathPrefix, "/") {
			names = append(names, pathPrefix+name)
		} else {
			names = append(names, pathPrefix+"/"+name)
		}
	}
	return names
}

// newUDBBackup ucloud udb backup
func newUDBBackup(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "List and manipulate backups of MySQL instance",
		Long:  "List and manipulate backups of MySQL instance",
	}
	cmd.AddCommand(newUDBBackupCreate(ctx))
	cmd.AddCommand(newUDBBackupList(ctx))
	cmd.AddCommand(newUDBBackupDelete(ctx))
	cmd.AddCommand(newUDBBackupGetDownloadURL(ctx))
	return cmd
}

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
			fmt.Fprintf(ctx.Out(), "udb[%s] backuped\n", *req.DBId)
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
					BackupBeginTime:  formatDateTime(ins.BackupTime),
					BackupEndTime:    formatDateTime(ins.BackupEndTime),
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

// newUDBBackupDelete ucloud udb backup delete
func newUDBBackupDelete(ctx *cli.Context) *cobra.Command {
	ids := []int{}
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDeleteUDBBackupRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete backups of MySQL instance",
		Long:    "Delete backups of MySQL instance",
		Example: "ucloud udb backup delete --backup-id 65534,65535",
		Run: func(c *cobra.Command, args []string) {
			for _, id := range ids {
				req.BackupId = sdk.Int(id)
				_, err := client.DeleteUDBBackup(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.Out(), "backup[%d] deleted\n", id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.IntSliceVar(&ids, "backup-id", nil, "Required. BackupID of backups to delete")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("backup-id")
	return cmd
}

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

// newUDBLog ucloud udb log
func newUDBLog(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "List and manipulate logs of MySQL instance",
		Long:  "List and manipulate logs of MySQL instance",
	}

	cmd.AddCommand(newUDBLogArchiveCreate(ctx))
	cmd.AddCommand(newUDBLogArchiveList(ctx))
	cmd.AddCommand(newUDBLogArchiveGetDownloadURL(ctx))
	cmd.AddCommand(newUDBLogArchiveDelete(ctx))

	return cmd
}

// newUDBLogArchiveCreate ucloud udb log archive create
func newUDBLogArchiveCreate(ctx *cli.Context) *cobra.Command {
	var udbID string
	var name, logType, beginTime, endTime string
	var common request.CommonBase
	cmd := &cobra.Command{
		Use:     "archive",
		Short:   "Archive the log of mysql as a compressed file",
		Long:    "Archive the log of mysql as a compressed file",
		Example: "ucloud mysql logs archive --name test.cli2 --udb-id udb-xxx/test.cli1 --log-type slow_query --begin-time 2019-02-23/15:30:00 --end-time 2019-02-24/15:31:00",
		Run: func(c *cobra.Command, args []string) {
			region := common.GetRegion()
			zone := common.GetZone()
			project := common.GetProjectId()
			udbID = ctx.PickResourceID(udbID)
			client := cli.NewServiceClient(ctx, udb.NewClient)
			if logType == "slow_query" {
				if beginTime == "" || endTime == "" {
					fmt.Fprintln(ctx.Out(), "Error. Both begin-time and end-time can not be empty")
					return
				}
				bt, err := time.Parse(dateTimeLayout, beginTime)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				et, err := time.Parse(dateTimeLayout, endTime)
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
				fmt.Fprintf(ctx.Out(), "mysql log archive[%s] created\n", name)
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
				fmt.Fprintf(ctx.Out(), "mysql log archive[%s] created\n", name)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&udbID, "udb-id", "", "Required. Resource ID of UDB instance which we fetch logs from")
	flags.StringVar(&name, "name", "", "Required. Name of compressed file")
	flags.StringVar(&logType, "log-type", "", "Required. Type of log to package. Accept values: slow_query, error")
	flags.StringVar(&beginTime, "begin-time", "", "Optional. Required when log-type is slow. For example 2019-01-02/15:04:05")
	flags.StringVar(&endTime, "end-time", "", "Optional. Required when log-type is slow. For example 2019-01-02/15:04:05")
	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)
	ctx.BindProjectID(cmd, &common)

	cmd.MarkFlagRequired("udb-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("log-type")

	command.SetFlagValues(cmd, "log-type", "slow_query", "error")
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sql", common.GetProjectId(), common.GetRegion(), common.GetZone())
	})
	return cmd
}

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
				bt, err := time.Parse(dateTimeLayout, beginTime)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.BeginTime = sdk.Int(int(bt.Unix()))
			}
			if endTime != "" {
				et, err := time.Parse(dateTimeLayout, endTime)
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
					fmt.Fprintln(ctx.Out(), "Error, log-type should be one of 'binlog', 'slow_query' or 'error'")
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
					CreateTime: formatDateTime(ins.BackupTime),
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

// newUDBLogArchiveGetDownloadURL ucloud udb log archive get-download-url
func newUDBLogArchiveGetDownloadURL(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBBinlogBackupURLRequest()
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Display url of an archive(log file)",
		Long:    "Display url of an archive(log file)",
		Example: "ucloud mysql logs download --udb-id udb-urixxx/test.cli1 --archive-id 35044",
		Run: func(c *cobra.Command, args []string) {
			*req.DBId = ctx.PickResourceID(*req.DBId)
			resp, err := client.DescribeUDBBinlogBackupURL(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintln(ctx.Out(), resp.BackupPath)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.BackupId = flags.Int("archive-id", 0, "Required. ArchiveID of archive to download")
	req.DBId = flags.String("udb-id", "", "Required. Resource ID of UDB which the archive belongs to")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("archive-id")
	cmd.MarkFlagRequired("udb-id")

	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// newUDBLogArchiveDelete ucloud udb log archive delete
func newUDBLogArchiveDelete(ctx *cli.Context) *cobra.Command {
	var ids []int
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDeleteUDBLogPackageRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete log archives(log files)",
		Long:    "Delete log archives(log files)",
		Example: "ucloud mysql logs delete --archive-id 35025",
		Run: func(c *cobra.Command, args []string) {
			for _, id := range ids {
				req.BackupId = sdk.Int(id)
				_, err := client.DeleteUDBLogPackage(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.Out(), "archive[%d] deleted\n", id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.IntSliceVar(&ids, "archive-id", nil, "Optional. ArchiveID of log archives to delete")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("archive-id")

	return cmd
}

// Copyright © 2018 NAME HERE tony.li@ucloud.cn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
)

//NewCmdUDBBackup ucloud udb backup
func NewCmdUDBBackup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "List and manipulate backups of MySQL instance",
		Long:  "List and manipulate backups of MySQL instance",
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUDBBackupCreate(out))
	cmd.AddCommand(NewCmdUDBBackupList(out))
	cmd.AddCommand(NewCmdUDBBackupDelete(out))
	cmd.AddCommand(NewCmdUDBBackupGetDownloadURL(out))
	return cmd
}

//NewCmdUDBBackupCreate ucloud udb backup create
func NewCmdUDBBackupCreate(out io.Writer) *cobra.Command {
	req := base.BizClient.NewBackupUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create backups for MySQL instance manually",
		Long:  "Create backups for MySQL instance manually",
		Run: func(c *cobra.Command, args []string) {
			*req.DBId = base.PickResourceID(*req.DBId)
			_, err := base.BizClient.BackupUDBInstance(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "udb[%s] backuped\n", *req.DBId)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBId = flags.String("udb-id", "", "Required. Resource ID of UDB instnace to backup")
	req.BackupName = flags.String("name", "", "Required. Name of backup")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)

	cmd.MarkFlagRequired("udb-id")
	cmd.MarkFlagRequired("name")

	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
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

//NewCmdUDBBackupList ucloud udb backup list
func NewCmdUDBBackupList(out io.Writer) *cobra.Command {
	var bpType, dbType, beginTime, endTime, backupID string
	bpTypeMap := map[string]int{
		"manual": 1,
		"auto":   0,
	}
	reverseBpTypeMap := map[int]string{
		1: "manual",
		0: "auto",
	}
	req := base.BizClient.NewDescribeUDBBackupRequest()
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
				*req.DBId = base.PickResourceID(*req.DBId)
			}
			if backupID != "" {
				id, err := strconv.Atoi(base.PickResourceID(backupID))
				if err != nil {
					base.HandleError(err)
					return
				}
				req.BackupId = &id
			}
			if beginTime != "" {
				bt, err := time.Parse("2006-01-02/15:04:05", beginTime)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.BeginTime = sdk.Int(int(bt.Unix()))
			}
			if endTime != "" {
				bt, err := time.Parse("2006-01-02/15:04:05", endTime)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.EndTime = sdk.Int(int(bt.Unix()))
			}
			resp, err := base.BizClient.DescribeUDBBackup(req)
			if err != nil {
				base.HandleError(err)
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
					BackupBeginTime:  base.FormatDateTime(ins.BackupTime),
					BackupEndTime:    base.FormatDateTime(ins.BackupEndTime),
				}
				list = append(list, row)
			}
			base.PrintList(list, out)
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

	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	bindOffset(req, flags)
	bindLimit(req, flags)

	flags.SetFlagValues("backup-type", "auto", "manual")
	flags.SetFlagValues("db-type", dbTypeList...)
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

//NewCmdUDBBackupDelete ucloud udb backup delete
func NewCmdUDBBackupDelete(out io.Writer) *cobra.Command {
	ids := []int{}
	req := base.BizClient.NewDeleteUDBBackupRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete backups of MySQL instance",
		Long:    "Delete backups of MySQL instance",
		Example: "ucloud udb backup delete --backup-id 65534,65535",
		Run: func(c *cobra.Command, args []string) {
			for _, id := range ids {
				req.BackupId = sdk.Int(id)
				_, err := base.BizClient.DeleteUDBBackup(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "backup[%d] deleted\n", id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.IntSliceVar(&ids, "backup-id", nil, "Required. BackupID of backups to delete")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)

	cmd.MarkFlagRequired("backup-id")
	return cmd
}

//NewCmdUDBBackupGetDownloadURL ucloud udb backup get-download-url
func NewCmdUDBBackupGetDownloadURL(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeUDBInstanceBackupURLRequest()
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Display download url of backup",
		Long:  "Display download url of backup",
		Run: func(c *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeUDBInstanceBackupURL(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintln(out, resp.BackupPath)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.BackupId = flags.Int("backup-id", -1, "Required. BackupID of backup to delete")
	req.DBId = flags.String("udb-id", "", "Required. Resource ID of udb which the backup belongs to")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)

	cmd.MarkFlagRequired("udb-id")
	cmd.MarkFlagRequired("backup-id")
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

//NewCmdUDBLog ucloud udb log
func NewCmdUDBLog() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "List and manipulate logs of MySQL instance",
		Long:  "List and manipulate logs of MySQL instance",
	}

	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUDBLogArchiveCreate(out))
	cmd.AddCommand(NewCmdUDBLogArchiveList(out))
	cmd.AddCommand(NewCmdUDBLogArchiveGetDownloadURL(out))
	cmd.AddCommand(NewCmdUDBLogArchiveDelete(out))

	return cmd
}

//NewCmdUDBLogArchiveCreate ucloud udb log archive create
func NewCmdUDBLogArchiveCreate(out io.Writer) *cobra.Command {
	var region, zone, project, udbID string
	var name, logType, beginTime, endTime string
	cmd := &cobra.Command{
		Use:     "archive",
		Short:   "Archive the log of mysql as a compressed file",
		Long:    "Archive the log of mysql as a compressed file",
		Example: "ucloud mysql logs archive --name test.cli2 --udb-id udb-xxx/test.cli1 --log-type slow_query --begin-time 2019-02-23/15:30:00 --end-time 2019-02-24/15:31:00",
		Run: func(c *cobra.Command, args []string) {
			udbID = base.PickResourceID(udbID)
			if logType == "slow_query" {
				if beginTime == "" || endTime == "" {
					fmt.Fprintln(out, "Error. Both begin-time and end-time can not be empty")
					return
				}
				bt, err := time.Parse(base.DateTimeLayout, beginTime)
				if err != nil {
					base.HandleError(err)
					return
				}
				et, err := time.Parse(base.DateTimeLayout, endTime)
				if err != nil {
					base.HandleError(err)
					return
				}

				req := base.BizClient.NewBackupUDBInstanceSlowLogRequest()
				req.BeginTime = sdk.Int(int(bt.Unix()))
				req.EndTime = sdk.Int(int(et.Unix()))
				req.DBId = &udbID
				req.BackupName = &name
				req.Region = &region
				req.ProjectId = &project

				_, err = base.BizClient.BackupUDBInstanceSlowLog(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "mysql log archive[%s] created\n", name)
			} else if logType == "error" {
				req := base.BizClient.NewBackupUDBInstanceErrorLogRequest()
				req.DBId = &udbID
				req.BackupName = &name
				req.Region = &region
				req.Zone = &zone
				req.ProjectId = &project

				_, err := base.BizClient.BackupUDBInstanceErrorLog(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "mysql log archive[%s] created\n", name)
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
	bindRegionS(&region, flags)
	bindZoneS(&zone, &region, flags)
	bindProjectIDS(&project, flags)

	cmd.MarkFlagRequired("udb-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("log-type")

	flags.SetFlagValues("log-type", "slow_query", "error")
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "sql", project, region, base.ConfigIns.Zone)
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

//NewCmdUDBLogArchiveList ucloud udb log archive list
func NewCmdUDBLogArchiveList(out io.Writer) *cobra.Command {
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
	req := base.BizClient.NewDescribeUDBLogPackageRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mysql log archives(log files)",
		Long:  "List mysql log archives(log files)",
		Run: func(c *cobra.Command, args []string) {
			if beginTime != "" {
				bt, err := time.Parse(base.DateTimeLayout, beginTime)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.BeginTime = sdk.Int(int(bt.Unix()))
			}
			if endTime != "" {
				et, err := time.Parse(base.DateTimeLayout, endTime)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.EndTime = sdk.Int(int(et.Unix()))
			}

			if *req.DBId != "" {
				*req.DBId = base.PickResourceID(*req.DBId)
			}

			for _, s := range logTypes {
				if v, ok := logTypeMap[s]; ok {
					req.Types = append(req.Types, v)
				} else {
					fmt.Fprintln(out, "Error, log-type should be one of 'binlog', 'slow_query' or 'error'")
				}
			}

			resp, err := base.BizClient.DescribeUDBLogPackage(req)
			if err != nil {
				base.HandleError(err)
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
					CreateTime: base.FormatDateTime(ins.BackupTime),
				}
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&logTypes, "log-type", nil, "Optional. Type of log. Accept Values: binlog, slow_query and error")
	req.DBId = flags.String("udb-id", "", "Optional. Resource ID of UDB instance which the listed logs belong to")
	flags.StringVar(&beginTime, "begin-time", "", "Optional. For example 2019-01-02/15:04:05")
	flags.StringVar(&endTime, "end-time", "", "Optional. For example 2019-01-02/15:04:05")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)
	bindLimit(req, flags)
	bindOffset(req, flags)

	flags.SetFlagValues("log-type", "binlog", "slow_query", "error")
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

//NewCmdUDBLogArchiveGetDownloadURL ucloud udb log archive get-download-url
func NewCmdUDBLogArchiveGetDownloadURL(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeUDBBinlogBackupURLRequest()
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Display url of an archive(log file)",
		Long:    "Display url of an archive(log file)",
		Example: "ucloud mysql logs download --udb-id udb-urixxx/test.cli1 --archive-id 35044",
		Run: func(c *cobra.Command, args []string) {
			*req.DBId = base.PickResourceID(*req.DBId)
			resp, err := base.BizClient.DescribeUDBBinlogBackupURL(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintln(out, resp.BackupPath)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.BackupId = flags.Int("archive-id", 0, "Required. ArchiveID of archive to download")
	req.DBId = flags.String("udb-id", "", "Required. Resource ID of UDB which the archive belongs to")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("archive-id")
	cmd.MarkFlagRequired("udb-id")

	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

//NewCmdUDBLogArchiveDelete ucloud udb log archive delete
func NewCmdUDBLogArchiveDelete(out io.Writer) *cobra.Command {
	var ids []int
	req := base.BizClient.NewDeleteUDBLogPackageRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete log archives(log files)",
		Long:    "Delete log archives(log files)",
		Example: "ucloud mysql logs delete --archive-id 35025",
		Run: func(c *cobra.Command, args []string) {
			for _, id := range ids {
				req.BackupId = sdk.Int(id)
				_, err := base.BizClient.DeleteUDBLogPackage(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "archive[%d] deleted\n", id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.IntSliceVar(&ids, "archive-id", nil, "Optional. ArchiveID of log archives to delete")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("archive-id")

	return cmd
}

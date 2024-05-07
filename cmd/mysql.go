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
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
)

var dbVersionList = []string{"mysql-5.1", "mysql-5.5", "mysql-5.6", "mysql-5.7", "percona-5.5", "percona-5.6", "percona-5.7", "mariadb-10.0"}
var dbDiskTypeList = []string{"normal", "sata_ssd", "pcie_ssd"}

var poller = base.NewSpoller(describeUdbByID, base.Cxt.GetWriter())

// NewCmdMysql ucloud mysql
func NewCmdMysql() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mysql",
		Short: "Manipulate MySQL on UCloud platform",
		Long:  "Manipulate MySQL on UCloud platform",
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdMysqlDB(out))
	cmd.AddCommand(NewCmdUDBConf())
	cmd.AddCommand(NewCmdUDBBackup())
	cmd.AddCommand(NewCmdUDBLog())
	return cmd
}

// NewCmdMysqlDB ucloud mysql db
func NewCmdMysqlDB(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manange MySQL instances",
		Long:  "Manange MySQL instances",
	}

	cmd.AddCommand(NewCmdUDBList(out))
	cmd.AddCommand(NewCmdMysqlCreate(out))
	cmd.AddCommand(NewCmdUDBDelete(out))
	cmd.AddCommand(NewCmdUDBStart(out))
	cmd.AddCommand(NewCmdUDBStop(out))
	cmd.AddCommand(NewCmdUDBRestart(out))
	cmd.AddCommand(NewCmdUDBResize(out))
	cmd.AddCommand(NewCmdUDBRestore(out))
	cmd.AddCommand(NewCmdUDBResetPassword(out))
	cmd.AddCommand(NewCmdUDBCreateSlave(out))
	cmd.AddCommand(NewCmdUDBPromoteSlave(out))
	// cmd.AddCommand(NewCmdUDBPromoteToHA(out))

	return cmd
}

// NewCmdMysqlCreate ucloud mysql create
func NewCmdMysqlCreate(out io.Writer) *cobra.Command {
	var confID, diskType string
	var backupID int
	var async bool
	req := base.BizClient.NewCreateUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create MySQL instance on UCloud platform",
		Long:  "Create MySQL instance on UCloud platform",
		Run: func(c *cobra.Command, args []string) {
			confID = base.PickResourceID(confID)
			id, err := strconv.Atoi(confID)
			if err != nil {
				base.HandleError(err)
				return
			}
			req.ParamGroupId = &id
			if len(*req.Name) < 6 {
				fmt.Fprintln(out, "Error, length of name shoud be larger than 5")
				return
			}
			if *req.DiskSpace > 3000 || *req.DiskSpace < 20 {
				fmt.Fprintln(out, "Error, disk-size-gb should be between 20 and 3000")
				return
			}
			if *req.MemoryLimit < 1 || *req.MemoryLimit > 128 {
				fmt.Fprintln(out, "Error, memory-size-gb should be between 1 and 128")
				return
			}
			if backupID != -1 {
				req.BackupId = &backupID
			}
			*req.MemoryLimit = *req.MemoryLimit * 1000
			switch diskType {
			case "normal":
				req.UseSSD = sdk.Bool(false)
			case "sata_ssd":
				req.UseSSD = sdk.Bool(true)
				req.SSDType = sdk.String("SATA")
			case "pcie_ssd":
				req.UseSSD = sdk.Bool(true)
				req.SSDType = sdk.String("PCI-E")
			default:
				if diskType != "" {
					req.UseSSD = sdk.Bool(true)
					req.SSDType = sdk.String(diskType)
				}
			}
			resp, err := base.BizClient.CreateUDBInstance(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			text := fmt.Sprintf("udb[%s] is initializing", resp.DBId)
			if async {
				fmt.Fprintf(out, "udb[%s] is initializing\n", resp.DBId)
			} else {
				poller.Spoll(resp.DBId, text, []string{status.UDB_RUNNING, status.UDB_FAIL})
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)
	req.DBTypeId = flags.String("version", "", "Required. Version of udb instance")
	req.Name = flags.String("name", "", "Required. Name of udb instance to create, at least 6 letters")
	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of configuration. see 'ucloud mysql conf list'")
	req.AdminUser = flags.String("admin-user-name", "root", "Optional. Name of udb instance's administrator")
	req.AdminPassword = flags.String("password", "", "Required. Password of udb instance's administrator")
	flags.IntVar(&backupID, "backup-id", -1, "Optional. BackupID of the backup which the newly created UDB instance will recover from if specified. See 'ucloud mysql backup list'")
	req.Port = flags.Int("port", 3306, "Optional. Port of udb instance")
	flags.StringVar(&diskType, "disk-type", "", "Optional. Setting this flag means using SSD disk. Accept values: 'normal','sata_ssd','pcie_ssd'")
	req.DiskSpace = flags.Int("disk-size-gb", 20, "Optional. Disk size of udb instance. From 20 to 3000 according to memory size. Unit GB")
	req.MemoryLimit = flags.Int("memory-size-gb", 1, "Optional. Memory size of udb instance. From 1 to 128. Unit GB")
	req.InstanceMode = flags.String("mode", "Normal", "Optional. Mode of udb instance. Normal or HA, HA means high-availability. Both the normal and high-availability versions can create master-slave synchronization for data redundancy and read/write separation. The high-availability version provides a dual-master hot standby architecture to avoid database unavailability due to downtime or hardware failure. One more thing. It does better job for master-slave synchronization and disaster recovery using the InnoDB engine")
	req.VPCId = flags.String("vpc-id", "", "Optional. Resource ID of VPC which the UDB to create belong to. See 'ucloud vpc list'")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Resource ID of subnet that the UDB to create belong to. See 'ucloud subnet list'")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the long-running operation to finish.")
	bindChargeType(req, flags)
	bindQuantity(req, flags)

	flags.SetFlagValues("version", dbVersionList...)
	flags.SetFlagValues("disk-type", dbDiskTypeList...)
	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("subnet-id", func() []string {
		return getAllSubnetIDNames(*req.VPCId, *req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("conf-id", func() []string {
		return getConfIDList(*req.DBTypeId, *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("conf-id")
	return cmd
}

// UDBMysqlRow 表格行
type UDBMysqlRow struct {
	Name       string
	ResourceID string
	Role       string
	Status     string
	Config     string
	Mode       string
	DiskType   string
	IP         string
	Group      string
	Zone       string
	VPC        string
	Subnet     string
	// CreateTime string
}

// NewCmdUDBList ucloud udb list
func NewCmdUDBList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List MySQL instances",
		Long:  "List MySQL instances",
		Run: func(c *cobra.Command, args []string) {
			if *req.DBId != "" {
				*req.DBId = base.PickResourceID(*req.DBId)
			}
			resp, err := base.BizClient.DescribeUDBInstance(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []UDBMysqlRow{}
			for _, ins := range resp.DataSet {
				row := UDBMysqlRow{}
				row.Name = ins.Name
				row.Zone = ins.Zone
				row.Role = ins.Role
				row.ResourceID = ins.DBId
				row.Group = ins.Tag
				row.VPC = ins.VPCId
				row.Subnet = ins.SubnetId
				row.IP = ins.VirtualIP
				row.Mode = ins.InstanceMode
				row.DiskType = ins.InstanceType
				row.Status = ins.State
				row.Config = fmt.Sprintf("%s|%dG|%dG", ins.DBTypeId, ins.MemoryLimit/1000, ins.DiskSpace)
				list = append(list, row)
				for _, slave := range ins.DataSet {
					row := UDBMysqlRow{}
					row.Name = slave.Name
					row.Zone = slave.Zone
					row.Role = fmt.Sprintf("\u2b91 %s", slave.Role)
					row.ResourceID = slave.DBId
					row.Group = slave.Tag
					row.VPC = slave.VPCId
					row.Subnet = slave.SubnetId
					row.IP = slave.VirtualIP
					row.Mode = slave.InstanceMode
					row.DiskType = slave.InstanceType
					row.Config = fmt.Sprintf("%s|%dG|%dG", slave.DBTypeId, slave.MemoryLimit/1000, slave.DiskSpace)
					row.Status = slave.State
					list = append(list, row)
				}
			}
			base.PrintList(list, out)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBId = flags.String("udb-id", "", "Optional. List the specified mysql")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)
	bindLimit(req, flags)
	bindOffset(req, flags)
	req.IncludeSlaves = flags.Bool("include-slaves", false, "Optional. When specifying the udb-id, whether to display its slaves together. Accept values:true, false")
	req.ClassType = sdk.String("sql")

	flags.SetFlagValues("include-slaves", "true", "false")
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// NewCmdUDBDelete ucloud udb delete
func NewCmdUDBDelete(out io.Writer) *cobra.Command {
	var idNames []string
	var yes bool
	req := base.BizClient.NewDeleteUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete MySQL instances by udb-id",
		Long:  "Delete MySQL instances by udb-id",
		Run: func(c *cobra.Command, args []string) {
			ok := base.Confirm(yes, "Are you sure you want to delete the udb(s)?")
			if !ok {
				return
			}
			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				any, err := describeUdbByID(id, nil)
				if err != nil {
					base.HandleError(err)
					continue
				}
				req.DBId = &id
				ins, ok := any.(*udb.UDBInstanceSet)
				if ok && ins.State == status.UDB_RUNNING {
					stopReq := base.BizClient.NewStopUDBInstanceRequest()
					stopReq.ProjectId = req.ProjectId
					stopReq.Region = req.Region
					stopReq.Zone = req.Zone
					stopReq.DBId = req.DBId
					stopUdbIns(stopReq, false, out)
				}
				_, err = base.BizClient.DeleteUDBInstance(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "udb[%s] deleted\n", idname)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to delete")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")

	cmd.MarkFlagRequired("udb-id")
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

// NewCmdUDBStop ucloud udb stop
func NewCmdUDBStop(out io.Writer) *cobra.Command {
	var idNames []string
	var async bool
	req := base.BizClient.NewStopUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop MySQL instances by udb-id",
		Long:  "Stop MySQL instances by udb-id",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				req.DBId = sdk.String(base.PickResourceID(idname))
				stopUdbIns(req, async, out)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to stop")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	req.ForceToKill = flags.Bool("force", false, "Optional. Stop UDB instances by force or not")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	cmd.MarkFlagRequired("udb-id")

	flags.SetFlagValues("force", "true", "false")
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList([]string{status.UDB_RUNNING}, "", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// NewCmdUDBStart ucloud udb start
func NewCmdUDBStart(out io.Writer) *cobra.Command {
	var async bool
	var idNames []string
	req := base.BizClient.NewStartUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start MySQL instances by udb-id",
		Long:  "Start MySQL instances by udb-id",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				req.DBId = &id
				_, err := base.BizClient.StartUDBInstance(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(out, "udb[%s] is starting\n", idname)
				} else {
					text := fmt.Sprintf("udb[%s] is starting", idname)
					poller.Spoll(*req.DBId, text, []string{status.UDB_RUNNING, status.UDB_FAIL})
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to start")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	cmd.MarkFlagRequired("udb-id")

	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList([]string{status.UDB_SHUTOFF}, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

// NewCmdUDBRestart ucloud udb restart
func NewCmdUDBRestart(out io.Writer) *cobra.Command {
	var async bool
	var idNames []string
	req := base.BizClient.NewRestartUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart MySQL instances by udb-id",
		Long:  "Restart MySQL instances by udb-id",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				req.DBId = &id
				_, err := base.BizClient.RestartUDBInstance(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(out, "udb[%s] is restarting\n", idname)
				} else {
					text := fmt.Sprintf("udb[%s] is restarting", idname)
					poller.Spoll(*req.DBId, text, []string{status.UDB_RUNNING, status.UDB_FAIL})
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to restart")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	cmd.MarkFlagRequired("udb-id")
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

// NewCmdUDBResize ucloud udb resize
func NewCmdUDBResize(out io.Writer) *cobra.Command {
	var diskTypes = []string{"normal", "sata_ssd", "pcie_ssd", "normal_volume", "sata_ssd_volume", "pcie_ssd_volume"}
	var async, yes bool
	var idNames []string
	var memory, disk int
	var diskType string
	req := base.BizClient.NewResizeUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Reszie MySQL instances, such as memory size, disk size and disk type",
		Long:  "Reszie MySQL instances, such as memory size, disk size and disk type",
		Run: func(c *cobra.Command, args []string) {
			if diskType != "" {
				switch diskType {
				case "normal":
					req.InstanceType = sdk.String("Normal")
				case "sata_ssd":
					req.InstanceType = sdk.String("SATA_SSD")
				case "pcie_ssd":
					req.InstanceType = sdk.String("PCIE_SSD")
				case "normal_volume":
					req.InstanceType = sdk.String("Normal_Volume")
				case "sata_ssd_volume":
					req.InstanceType = sdk.String("SATA_SSD_Volume")
				case "pcie_ssd_volume":
					req.InstanceType = sdk.String("PCIE_SSD_Volume")
				default:
					req.InstanceType = &diskType
				}
			}

			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				req.DBId = &id
				any, err := describeUdbByID(id, nil)
				if err != nil {
					base.HandleError(err)
					continue
				}

				ins, ok := any.(*udb.UDBInstanceSet)
				if !ok {
					continue
				}

				if memory != 0 {
					req.MemoryLimit = sdk.Int(memory * 1000)
				} else {
					req.MemoryLimit = &ins.MemoryLimit
				}
				if disk != 0 {
					req.DiskSpace = &disk
				} else {
					req.DiskSpace = &ins.DiskSpace
				}

				if ins.State == status.UDB_RUNNING {
					ok := base.Confirm(yes, fmt.Sprintf("Need to shut down udb[%s] before upgrading, whether to continue?", idname))
					if !ok {
						continue
					}
					stopReq := base.BizClient.NewStopUDBInstanceRequest()
					stopReq.ProjectId = req.ProjectId
					stopReq.Region = req.Region
					stopReq.Zone = req.Zone
					stopReq.DBId = req.DBId
					stopUdbIns(stopReq, false, out)
				}
				_, err = base.BizClient.ResizeUDBInstance(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(out, "udb[%s] is resizing\n", idname)
				} else {
					text := fmt.Sprintf("udb[%s] is resizing", idname)
					poller.Spoll(*req.DBId, text, []string{status.UDB_RUNNING, status.UDB_SHUTOFF, status.UDB_FAIL, status.UDB_UPGRADE_FAIL})
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to restart")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	flags.IntVar(&memory, "memory-size-gb", 0, "Optional. Memory size of udb instance. From 1 to 128. Unit GB")
	flags.IntVar(&disk, "disk-size-gb", 0, "Optional. Disk size of udb instance. From 20 to 3000 according to memory size. Unit GB. Step 10GB")
	flags.StringVar(&diskType, "disk-type", "", fmt.Sprintf("Optional. Disk type of udb instance. Accept values:%s", strings.Join(diskTypes, ", ")))
	req.StartAfterUpgrade = flags.Bool("start-after-upgrade", true, "Optional. Automatic start the UDB instances after upgrade")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation")

	flags.SetFlagValues("disk-type", diskTypes...)
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udb-id")

	return cmd
}

// NewCmdUDBResetPassword ucloud udb reset-password
func NewCmdUDBResetPassword(out io.Writer) *cobra.Command {
	var idNames []string
	req := base.BizClient.NewModifyUDBInstancePasswordRequest()
	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset password of MySQL instances",
		Long:  "Reset password of MySQL instances",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				req.DBId = &id
				_, err := base.BizClient.ModifyUDBInstancePassword(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "udb[%s]'s password modified\n", idname)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to reset password")
	req.Password = flags.String("password", "", "Required. New password")
	bindProjectID(req, flags)
	bindRegion(req, flags)
	bindZone(req, flags)

	cmd.MarkFlagRequired("udb-id")
	cmd.MarkFlagRequired("password")

	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// NewCmdUDBRestore ucloud udb restore
func NewCmdUDBRestore(out io.Writer) *cobra.Command {
	var datetime, diskType string
	var async bool
	req := base.BizClient.NewCreateUDBInstanceByRecoveryRequest()
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Create MySQL instance and restore the newly created db to the specified DB at a specified point in time",
		Long:  "Create MySQL instance and restore the newly created db to the specified DB at a specified point in time",
		Run: func(c *cobra.Command, args []string) {
			t, err := time.Parse(time.RFC3339, datetime)
			if err != nil {
				base.HandleError(err)
				return
			}
			req.RecoveryTime = sdk.Int(int(t.Unix()))
			req.SrcDBId = sdk.String(base.PickResourceID(*req.SrcDBId))
			if diskType == "" {
				any, err := describeUdbByID(*req.SrcDBId, nil)
				if err != nil {
					base.HandleError(err)
					return
				}
				ins, ok := any.(*udb.UDBInstanceSet)
				if !ok {
					fmt.Fprintln(out, fmt.Sprintf("fetch udb[%s] instance", *req.SrcDBId))
				}
				req.UseSSD = &ins.UseSSD
			} else if diskType == "normal" {
				req.UseSSD = sdk.Bool(false)
			} else if diskType == "ssd" {
				req.UseSSD = sdk.Bool(true)
			}
			resp, err := base.BizClient.CreateUDBInstanceByRecovery(req)
			if async {
				fmt.Fprintf(out, "udb[%s] is restorting from udb[%s] at time point %s", resp.DBId, *req.SrcDBId, datetime)
			} else {
				text := fmt.Sprintf("udb[%s] is restorting from udb[%s] at time point %s", resp.DBId, *req.SrcDBId, datetime)
				poller.Spoll(resp.DBId, text, []string{status.UDB_RUNNING, status.UDB_RECOVER_FAIL, status.UDB_FAIL})
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = flags.String("name", "", "Required. Name of UDB instance to create")
	req.SrcDBId = flags.String("src-udb-id", "", "Required. Resource ID of source UDB")
	flags.StringVar(&datetime, "restore-to-time", "", "Required. The date and time to restore the DB to. Value must be a time in Universal Coordinated Time (UTC) format.Example: 2019-02-23T23:45:00Z")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	flags.StringVar(&diskType, "disk-type", "", "Optional. Disk type. The default is to be consistent with the source database. Accept values: normal, ssd")
	bindChargeType(req, flags)
	bindQuantity(req, flags)
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("src-udb-id")
	cmd.MarkFlagRequired("restore-to-time")

	flags.SetFlagValues("disk-type", "noraml", "ssd")
	flags.SetFlagValuesFunc("src-udb-id", func() []string {
		return getUDBIDList(nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// NewCmdUDBCreateSlave ucloud udb create-slave
func NewCmdUDBCreateSlave(out io.Writer) *cobra.Command {
	var diskType string
	var async bool
	req := base.BizClient.NewCreateUDBSlaveRequest()
	cmd := &cobra.Command{
		Use:   "create-slave",
		Short: "Create slave database",
		Long:  "Create slave database",
		Run: func(c *cobra.Command, args []string) {
			*req.SrcId = base.PickResourceID(*req.SrcId)
			switch diskType {
			case "normal":
				req.UseSSD = sdk.Bool(false)
			case "sata_ssd":
				req.UseSSD = sdk.Bool(true)
				req.SSDType = sdk.String("SATA")
			case "pcie_ssd":
				req.UseSSD = sdk.Bool(true)
				req.SSDType = sdk.String("PCI-E")
			}
			*req.MemoryLimit *= 1000
			resp, err := base.BizClient.CreateUDBSlave(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if async {
				fmt.Fprintf(out, "udb[%s] is initializing\n", resp.DBId)
			} else {
				poller.Spoll(resp.DBId, fmt.Sprintf("udb[%s] is initializing", resp.DBId), []string{status.UDB_RUNNING, status.UDB_FAIL})
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.SrcId = flags.String("master-udb-id", "", "Required. Resource ID of master UDB instance")
	req.Name = flags.String("name", "", "Required. Name of the slave DB to create")
	req.Port = flags.Int("port", 3306, "Optional. Port of the slave db service")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	flags.StringVar(&diskType, "disk-type", "Normal", fmt.Sprintf("Optional. Setting this flag means using SSD disk. Accept values: %s", strings.Join(dbDiskTypeList, ", ")))
	req.MemoryLimit = flags.Int("memory-size-gb", 1, "Optional. Memory size of udb instance. From 1 to 128. Unit GB")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the long-running operation to finish")
	req.IsLock = flags.Bool("is-lock", false, "Optional. Lock master DB or not")

	cmd.MarkFlagRequired("master-udb-id")
	cmd.MarkFlagRequired("name")

	flags.SetFlagValues("disk-type", dbDiskTypeList...)
	flags.SetFlagValuesFunc("master-udb-id", func() []string {
		return getUDBIDList(nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

// NewCmdUDBPromoteSlave ucloud udb promote-slave
func NewCmdUDBPromoteSlave(out io.Writer) *cobra.Command {
	var ids []string
	req := base.BizClient.NewPromoteUDBSlaveRequest()
	cmd := &cobra.Command{
		Use:   "promote-slave",
		Short: "Promote slave db to master",
		Long:  "Promote slave db to master",
		Run: func(c *cobra.Command, args []string) {
			for _, id := range ids {
				req.DBId = sdk.String(id)
				_, err := base.BizClient.PromoteUDBSlave(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "udb[%s] was promoted\n", *req.DBId)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, "udb-id", nil, "Required. Resource ID of slave db to promote")
	req.IsForce = flags.Bool("is-force", false, "Optional. Force to promote slave db or not. If the slave db falls behind, the force promote may lose some data")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("udb-id")

	return cmd
}

// NewCmdUDBPromoteToHA ucloud udb promote-to-ha 低频操作 暂不开放
func NewCmdUDBPromoteToHA(out io.Writer) *cobra.Command {
	var idNames []string
	req := base.BizClient.NewPromoteUDBInstanceToHARequest()
	cmd := &cobra.Command{
		Use:   "promote-to-ha",
		Short: "Promote db of normal mode to high availability db. ",
		Long:  "Promote db of normal mode to high availability db",
		Run: func(c *cobra.Command, args []string) {
			for _, idname := range idNames {
				id := base.PickResourceID(idname)
				req.DBId = &id
				_, err := base.BizClient.PromoteUDBInstanceToHA(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				poller.Spoll(id, fmt.Sprintf("udb[%s] is synchronizing data", id), []string{status.UDB_TOBE_SWITCH, status.UDB_FAIL})
				any, err := describeUdbByID(id, nil)
				if err != nil {
					fmt.Fprintf(out, "udb[%s] promoted failed, please contact technical support; %v\n", idname, err)
					continue
				}
				ins, ok := any.(*udb.UDBInstanceSet)
				if !ok {
					fmt.Fprintf(out, "udb[%s] promoted failed, please contact technical support. \n", idname)
					continue
				}
				if ins.State != status.UDB_TOBE_SWITCH {
					fmt.Fprintf(out, "udb[%s] promoted failed, please contact technical support. udb[%s]'s status:%s\n", idname, idname, ins.State)
					continue
				}
				switchReq := base.BizClient.NewSwitchUDBInstanceToHARequest()
				switchReq.DBId = &id
				switchReq.Region = req.Region
				switchReq.ProjectId = req.ProjectId
				switchReq.ChargeType = &ins.ChargeType
				switchReq.Quantity = sdk.String("0")
				switchReq.Zone = &base.ConfigIns.Zone
				switchResp, err := base.BizClient.SwitchUDBInstanceToHA(switchReq)
				if err != nil {
					fmt.Fprintf(out, "udb[%s] promoted failed, please contact technical support; %v\n", idname, err)
					continue
				}
				poller.Spoll(switchResp.DBId, fmt.Sprintf("udb[%s] is switching to high availability mode", switchResp.DBId), []string{status.UDB_RUNNING, status.UDB_FAIL})
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindProjectID(req, flags)
	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to be promoted as high availability mode")

	cmd.MarkFlagRequired("udb-id")
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "", *req.ProjectId, *req.Region, "")
	})
	return cmd
}

func stopUdbIns(req *udb.StopUDBInstanceRequest, async bool, out io.Writer) {
	_, err := base.BizClient.StopUDBInstance(req)
	if err != nil {
		base.HandleError(err)
		return
	}
	text := fmt.Sprintf("udb[%s] is stopping", *req.DBId)
	if async {
		fmt.Fprintln(out, text)
	} else {
		poller.Spoll(*req.DBId, text, []string{status.UDB_SHUTOFF, status.UDB_FAIL})
	}
}

func getUDBIDList(states []string, dbType, project, region, zone string) []string {
	udbs, err := getUDBList(states, dbType, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, db := range udbs {
		list = append(list, fmt.Sprintf("%s/%s", db.DBId, db.Name))
	}
	return list
}

func getUDBList(states []string, dbType, project, region, zone string) ([]udb.UDBInstanceSet, error) {
	req := base.BizClient.NewDescribeUDBInstanceRequest()
	if dbType == "" {
		dbType = "sql"
	}
	req.ClassType = &dbType
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	list := []udb.UDBInstanceSet{}
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := base.BizClient.DescribeUDBInstance(req)
		if err != nil {
			return nil, err
		}
		for _, ins := range resp.DataSet {
			if states != nil {
				for _, s := range states {
					if s == ins.State {
						list = append(list, ins)
					}
				}
			} else {
				list = append(list, ins)
			}
		}
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return list, nil
}

func describeUdbByID(udbID string, commonBase *request.CommonBase) (interface{}, error) {
	req := base.BizClient.NewDescribeUDBInstanceRequest()
	if commonBase != nil {
		req.CommonBase = *commonBase
	}
	req.DBId = sdk.String(udbID)
	resp, err := base.BizClient.DescribeUDBInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) < 1 {
		return nil, fmt.Errorf("udb[%s] may not exist", udbID)
	}
	return &resp.DataSet[0], nil
}

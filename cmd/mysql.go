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
	pflag "github.com/spf13/pflag"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
)

var dbVersionList = []string{"mysql-5.7", "mysql-8.0", "mysql-8.4", "percona-5.7"}
var dbDiskTypeList = []string{"normal", "sata_ssd", "pcie_ssd"}
var dbStorageClassList = []string{"CLOUD_RSSD"}
var dbSpecClassList = []string{"O", "O2"}
var dbMachineTypeList = []string{
	"o.mysql2m.small",    // 1C2G
	"o.mysql2m.medium",   // 2C4G
	"o.mysql2m.xlarge",   // 4C8G
	"o.mysql2m.2xlarge",  // 8C16G
	"o.mysql2m.4xlarge",  // 16C32G
	"o.mysql2m.8xlarge",  // 32C64G
	"o.mysql2m.12xlarge", // 48C96G
	"o.mysql2m.16xlarge", // 64C128G
	"o.mysql4m.medium",   // 2C8G
	"o.mysql4m.xlarge",   // 4C16G
	"o.mysql4m.2xlarge",  // 8C32G
	"o.mysql4m.4xlarge",  // 16C64G
	"o.mysql4m.8xlarge",  // 32C128G
	"o.mysql4m.16xlarge", // 64C256G
	"o.mysql8m.medium",   // 2C16G
	"o.mysql8m.xlarge",   // 4C32G
	"o.mysql8m.2xlarge",  // 8C64G
	"o.mysql8m.4xlarge",  // 16C128G
	"o.mysql8m.8xlarge",  // 32C256G
	"o.mysql8m.16xlarge", // 64C512G
}

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

// getDefaultParamGroupID 通过 ListUDBParamTemplate 获取指定 DB 版本的默认配置模板 ID
// templateType=0 表示默认参数模板
func getDefaultParamGroupID(dbVersion, project, region, zone string) (int, error) {
	params := map[string]interface{}{
		"Action":    "ListUDBParamTemplate",
		"Region":    region,
		"Zone":      zone,
		"ProjectId": project,
		"DBVersion": dbVersion,
	}
	req := base.BizClient.UAccountClient.NewGenericRequest()
	if err := req.SetPayload(params); err != nil {
		return 0, fmt.Errorf("set payload: %w", err)
	}
	resp, err := base.BizClient.UAccountClient.GenericInvoke(req)
	if err != nil {
		return 0, fmt.Errorf("call ListUDBParamTemplate: %w", err)
	}
	dataSet, ok := resp.GetPayload()["DataSet"].([]interface{})
	if !ok || len(dataSet) == 0 {
		return 0, fmt.Errorf("no param template found for version %s in %s/%s", dbVersion, region, zone)
	}
	// 取第一个默认模板
	m, _ := dataSet[0].(map[string]interface{})
	id, _ := m["Id"].(float64)
	return int(id), nil
}

// listParamTemplates 通过 ListUDBParamTemplate 获取指定版本的参数模板列表，用于自动补全
func listParamTemplates(dbVersion, project, region, zone string) []string {
	params := map[string]interface{}{
		"Action":    "ListUDBParamTemplate",
		"Region":    region,
		"Zone":      zone,
		"ProjectId": project,
		"DBVersion": dbVersion,
	}
	req := base.BizClient.UAccountClient.NewGenericRequest()
	if err := req.SetPayload(params); err != nil {
		return nil
	}
	resp, err := base.BizClient.UAccountClient.GenericInvoke(req)
	if err != nil {
		return nil
	}
	dataSet, ok := resp.GetPayload()["DataSet"].([]interface{})
	if !ok {
		return nil
	}
	var list []string
	for _, item := range dataSet {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := m["Id"].(float64)
		name, _ := m["Name"].(string)
		list = append(list, fmt.Sprintf("%d/%s", int(id), name))
	}
	return list
}

// NewCmdMysqlCreate ucloud mysql create
func NewCmdMysqlCreate(out io.Writer) *cobra.Command {
	var region, zone, projectID string
	var confID string
	var backupID int
	var async bool
	var labels []string
	var name, password, version, machineType, storageClass, specClass string
	var port, diskSpace int
	var chargeType string
	var quantity int
	var mode, vpcID, subnetID, backupZone string
	var backupCount, backupTime, backupDuration int
	var disableSemisync bool
	var tag, dbSubVersion, alarmTemplateID, backupURL string
	var caseSensitivity, semisyncFlag int
	var couponID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create MySQL instance on UCloud platform",
		Long:  "Create MySQL instance on UCloud platform",
		Run: func(c *cobra.Command, args []string) {
			if len(name) < 6 {
				fmt.Fprintln(out, "Error: name must be at least 6 characters")
				return
			}
			if diskSpace < 20 || diskSpace > 32000 {
				fmt.Fprintln(out, "Error: disk-size-gb must be between 20 and 32000")
				return
			}

			// ParamGroupId: 用户传了就用，没传则自动获取默认模板
			var paramGroupID int
			if c.Flags().Changed("param-group-id") {
				confID = base.PickResourceID(confID)
				id, err := strconv.Atoi(confID)
				if err != nil {
					base.HandleError(fmt.Errorf("invalid param-group-id: %w", err))
					return
				}
				paramGroupID = id
			} else {
				id, err := getDefaultParamGroupID(version, projectID, region, zone)
				if err != nil {
					base.HandleError(err)
					return
				}
				paramGroupID = id
			}

			params := map[string]interface{}{
				"Action":             "CreateUDBMySQLInstance",
				"Region":             region,
				"Zone":               zone,
				"Name":               name,
				"AdminPassword":      password,
				"DBTypeId":           version,
				"Port":               port,
				"DiskSpace":          diskSpace,
				"ParamGroupId":       paramGroupID,
				"MachineType":        machineType,
				"StorageClass":       storageClass,
				"SpecificationClass": specClass,
			}
			if projectID != "" {
				params["ProjectId"] = projectID
			}

			if c.Flags().Changed("charge-type") {
				params["ChargeType"] = chargeType
			}
			if c.Flags().Changed("quantity") {
				params["Quantity"] = quantity
			}
			if c.Flags().Changed("mode") {
				params["InstanceMode"] = mode
			}
			if c.Flags().Changed("vpc-id") {
				params["VPCId"] = vpcID
			}
			if c.Flags().Changed("subnet-id") {
				params["SubnetId"] = subnetID
			}
			if c.Flags().Changed("backup-zone") {
				params["BackupZone"] = backupZone
			}
			if c.Flags().Changed("backup-count") {
				params["BackupCount"] = backupCount
			}
			if c.Flags().Changed("backup-time") {
				params["BackupTime"] = backupTime
			}
			if c.Flags().Changed("backup-duration") {
				params["BackupDuration"] = backupDuration
			}
			if c.Flags().Changed("backup-id") {
				params["BackupId"] = backupID
			}
			if c.Flags().Changed("disable-semisync") {
				params["DisableSemisync"] = disableSemisync
			}
			if c.Flags().Changed("tag") {
				params["Tag"] = tag
			}
			if c.Flags().Changed("db-sub-version") {
				params["DBSubVersion"] = dbSubVersion
			}
			if c.Flags().Changed("case-sensitivity") {
				params["CaseSensitivityParam"] = caseSensitivity
			}
			if c.Flags().Changed("alarm-template-id") {
				params["AlarmTemplateId"] = alarmTemplateID
			}
			if c.Flags().Changed("backup-url") {
				params["BackupURL"] = backupURL
			}
			if c.Flags().Changed("semisync-flag") {
				params["SemisyncFlag"] = semisyncFlag
			}
			if c.Flags().Changed("coupon-id") {
				params["CouponId"] = couponID
			}

			for i, l := range labels {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) == 2 {
					params[fmt.Sprintf("Labels.%d.Key", i)] = parts[0]
					params[fmt.Sprintf("Labels.%d.Value", i)] = parts[1]
				}
			}

			req := base.BizClient.UAccountClient.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				base.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			resp, err := base.BizClient.UAccountClient.GenericInvoke(req)
			if err != nil {
				base.HandleError(err)
				return
			}

			dbID, _ := resp.GetPayload()["DBId"].(string)
			if dbID == "" {
				fmt.Fprintln(out, "Error: empty DBId in response")
				return
			}
			if async {
				fmt.Fprintf(out, "udb[%s] is initializing\n", dbID)
			} else {
				text := fmt.Sprintf("udb[%s] is initializing", dbID)
				poller.Spoll(dbID, text, []string{status.UDB_RUNNING, status.UDB_FAIL})
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	// Required flags
	flags.StringVar(&name, "name", "", "Required. Instance name, at least 6 characters")
	flags.StringVar(&password, "password", "", "Required. Admin password")
	flags.StringVar(&version, "version", "", "Required. DB version. Options: mysql-5.7, mysql-8.0, mysql-8.4, percona-5.7")
	flags.StringVar(&machineType, "machine-type", "", "Required. Machine type ID, e.g. o.mysql2m.xlarge for 4C8G. See ListUDBMachineType API")

	// Optional flags
	bindRegionS(&region, flags)
	bindZoneS(&zone, &region, flags)
	bindProjectIDS(&projectID, flags)
	flags.IntVar(&port, "port", 3306, "Optional. Port, default 3306")
	flags.IntVar(&diskSpace, "disk-size-gb", 20, "Optional. Disk size (GiB), 20-32000, default 20")
	flags.StringVar(&storageClass, "storage-class", "CLOUD_RSSD", "Optional. Storage class: CLOUD_RSSD")
	flags.StringVar(&specClass, "spec-class", "O", "Optional. Spec class: O(NVMe) / O2")

	flags.StringVar(&confID, "param-group-id", "", "Optional. Param group ID. Auto-fetched if omitted. See 'ucloud mysql conf list'")
	flags.StringVar(&chargeType, "charge-type", "Month", "Optional. Year / Month / Dynamic")
	flags.IntVar(&quantity, "quantity", 1, "Optional. Purchase duration")
	flags.IntVar(&backupID, "backup-id", -1, "Optional. Restore from backup ID")
	flags.StringVar(&mode, "mode", "Normal", "Optional. Normal / HA")
	flags.StringVar(&vpcID, "vpc-id", "", "Optional. VPC ID. See 'ucloud vpc list'")
	flags.StringVar(&subnetID, "subnet-id", "", "Optional. Subnet ID. See 'ucloud subnet list'")
	flags.StringVar(&backupZone, "backup-zone", "", "Optional. Backup zone for cross-AZ HA")
	flags.IntVar(&backupCount, "backup-count", 7, "Optional. Weekly backup count, default 7")
	flags.IntVar(&backupTime, "backup-time", 1, "Optional. Backup start hour (0-23), default 1")
	flags.IntVar(&backupDuration, "backup-duration", 24, "Optional. Backup interval hours, default 24")
	flags.BoolVar(&disableSemisync, "disable-semisync", false, "Optional. Enable async HA")
	flags.StringVar(&tag, "tag", "", "Optional. Business group name")
	flags.StringVar(&dbSubVersion, "db-sub-version", "", "Optional. MySQL minor version")
	flags.IntVar(&caseSensitivity, "case-sensitivity", -1, "Optional. 0=case-sensitive, 1=insensitive (MySQL 8.0 only)")
	flags.StringVar(&alarmTemplateID, "alarm-template-id", "", "Optional. Alarm template ID")
	flags.StringVar(&backupURL, "backup-url", "", "Optional. US3 backup download URL")
	flags.IntVar(&semisyncFlag, "semisync-flag", 0, "Optional. 1=enable semi-sync, 2=disable, 0=default(enable)")
	flags.StringSliceVar(&labels, "label", nil, "Optional. Resource label, format: key=value, repeatable")
	flags.StringVar(&couponID, "coupon-id", "", "Optional. Coupon ID")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for creation to finish")

	flags.SetFlagValues("version", dbVersionList...)
	flags.SetFlagValues("storage-class", dbStorageClassList...)
	flags.SetFlagValues("spec-class", dbSpecClassList...)
	flags.SetFlagValues("charge-type", "Month", "Dynamic", "Year")
	flags.SetFlagValues("mode", "Normal", "HA")

	flags.SetFlagValuesFunc("vpc-id", func() []string {
		return getAllVPCIdNames(projectID, region)
	})
	flags.SetFlagValuesFunc("subnet-id", func() []string {
		return getAllSubnetIDNames(vpcID, projectID, region)
	})
	flags.SetFlagValuesFunc("param-group-id", func() []string {
		return listParamTemplates(version, projectID, region, zone)
	})
	flags.SetFlagValues("machine-type", dbMachineTypeList...)

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("machine-type")

	// 自定义 usage，突出必填参数
	requiredFlags := []string{"name", "password", "version", "machine-type"}
	cmd.SetUsageFunc(func(c *cobra.Command) error {
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintf(out, "  %s [flags]\n\n", c.CommandPath())
		fmt.Fprintln(out, "★ Required flags (must be provided):")
		for _, name := range requiredFlags {
			f := c.Flags().Lookup(name)
			if f != nil {
				fmt.Fprintf(out, "  --%-20s %s\n", f.Name, f.Usage)
			}
		}
		fmt.Fprintln(out, "\nOptional flags:")
		c.Flags().VisitAll(func(f *pflag.Flag) {
			for _, req := range requiredFlags {
				if f.Name == req {
					return
				}
			}
			defVal := ""
			if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" && f.DefValue != "[]" {
				defVal = fmt.Sprintf(" (default %s)", f.DefValue)
			}
			fmt.Fprintf(out, "  --%-20s %s%s\n", f.Name, f.Usage, defVal)
		})
		return nil
	})

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

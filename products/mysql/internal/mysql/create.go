package mysql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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

// getDefaultParamGroupID 通过 ListUDBParamTemplate 获取指定 DB 版本的默认配置模板 ID
// ListUDBParamTemplate请求体 不传TemplateType即为获取默认参数模板
func getDefaultParamGroupID(ctx *cli.Context, dbVersion, project, region, zone string) (int, error) {
	params := map[string]interface{}{
		"Action":    "ListUDBParamTemplate",
		"Region":    region,
		"Zone":      zone,
		"ProjectId": project,
		"DBVersion": dbVersion,
	}
	client := cli.NewServiceClient(ctx, uaccount.NewClient)
	req := client.NewGenericRequest()
	if err := req.SetPayload(params); err != nil {
		return 0, fmt.Errorf("set payload: %w", err)
	}
	resp, err := client.GenericInvoke(req)
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
func listParamTemplates(ctx *cli.Context, dbVersion, project, region, zone string) []string {
	params := map[string]interface{}{
		"Action":    "ListUDBParamTemplate",
		"Region":    region,
		"Zone":      zone,
		"ProjectId": project,
		"DBVersion": dbVersion,
	}
	client := cli.NewServiceClient(ctx, uaccount.NewClient)
	req := client.NewGenericRequest()
	if err := req.SetPayload(params); err != nil {
		return nil
	}
	resp, err := client.GenericInvoke(req)
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

// newCreate ucloud mysql create
func newCreate(ctx *cli.Context) *cobra.Command {
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
	var common request.CommonBase

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create MySQL instance on UCloud platform",
		Long:  "Create MySQL instance on UCloud platform",
		Run: func(c *cobra.Command, args []string) {
			region := common.GetRegion()
			zone := common.GetZone()
			projectID := common.GetProjectId()
			if len(name) < 6 {
				ctx.HandleError(fmt.Errorf("name must be at least 6 characters"))
				return
			}
			if diskSpace < 20 || diskSpace > 32000 {
				ctx.HandleError(fmt.Errorf("disk-size-gb must be between 20 and 32000"))
				return
			}

			// ParamGroupId: 用户传了就用，没传则自动获取默认模板
			var paramGroupID int
			if c.Flags().Changed("param-group-id") {
				confID = ctx.PickResourceID(confID)
				id, err := strconv.Atoi(confID)
				if err != nil {
					ctx.HandleError(fmt.Errorf("invalid param-group-id: %w", err))
					return
				}
				paramGroupID = id
			} else {
				id, err := getDefaultParamGroupID(ctx, version, projectID, region, zone)
				if err != nil {
					ctx.HandleError(err)
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
				"ChargeType":         chargeType,
				"Quantity":           quantity,
				"InstanceMode":       mode,
				"BackupCount":        backupCount,
				"BackupTime":         backupTime,
				"BackupDuration":     backupDuration,
				"DisableSemisync":    disableSemisync,
				"SemisyncFlag":       semisyncFlag,
			}
			if projectID != "" {
				params["ProjectId"] = projectID
			}

			// 以下为可选参数，仅在用户显式指定时下发
			if c.Flags().Changed("vpc-id") {
				params["VPCId"] = vpcID
			}
			if c.Flags().Changed("subnet-id") {
				params["SubnetId"] = subnetID
			}
			if c.Flags().Changed("backup-zone") {
				params["BackupZone"] = backupZone
			}
			if c.Flags().Changed("backup-id") {
				params["BackupId"] = backupID
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
			if c.Flags().Changed("coupon-id") {
				params["CouponId"] = couponID
			}

			idx := 0
			for _, l := range labels {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) == 2 {
					params[fmt.Sprintf("Labels.%d.Key", idx)] = parts[0]
					params[fmt.Sprintf("Labels.%d.Value", idx)] = parts[1]
					idx++
				}
			}

			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			resp, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			dbID, _ := resp.GetPayload()["DBId"].(string)
			if dbID == "" {
				ctx.HandleError(fmt.Errorf("empty DBId in response"))
				return
			}
			w := ctx.ProgressWriter()
			if async {
				fmt.Fprintf(w, "udb[%s] is initializing\n", dbID)
			} else {
				text := fmt.Sprintf("udb[%s] is initializing", dbID)
				ctx.PollerTo(w, describeUdbByID(ctx)).Spoll(dbID, text, []string{UDB_RUNNING, UDB_FAIL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: dbID, Action: "create", Status: "Initializing"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	// Required flags
	flags.StringVar(&name, "name", "", "Required. Instance name, at least 6 characters")
	flags.StringVar(&password, "password", "", "Required. Admin password")
	flags.StringVar(&version, "version", "", "Required. DB version. Options: mysql-5.7, mysql-8.0, mysql-8.4, percona-5.7")
	flags.StringVar(&machineType, "machine-type", "", "Required. Machine type ID, e.g. o.mysql2m.xlarge for 4C8G. See 'ucloud mysql db list-machine-type'")

	// Optional flags
	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)
	ctx.BindProjectID(cmd, &common)
	flags.IntVar(&port, "port", 3306, "Optional. Port, default 3306")
	flags.IntVar(&diskSpace, "disk-size-gb", 20, "Optional. Disk size (GiB), 20-32000, default 20")
	flags.StringVar(&storageClass, "storage-class", "CLOUD_RSSD", "Optional. Storage class: CLOUD_RSSD")
	flags.StringVar(&specClass, "spec-class", "O", "Optional. Spec class: O(NVMe) / O2")

	flags.StringVar(&confID, "param-group-id", "", "Optional. Param group ID. Auto-fetched if omitted. See 'ucloud mysql conf list'")
	flags.StringVar(&chargeType, "charge-type", "Month", "Optional. Year / Month / Dynamic")
	flags.IntVar(&quantity, "quantity", 1, "Optional. Purchase duration")
	flags.IntVar(&backupID, "backup-id", -1, "Optional. Restore from backup ID")
	flags.StringVar(&mode, "mode", "HA", "Optional. Normal / HA")
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

	command.SetFlagValues(cmd, "version", dbVersionList...)
	command.SetFlagValues(cmd, "storage-class", dbStorageClassList...)
	command.SetFlagValues(cmd, "spec-class", dbSpecClassList...)
	command.SetFlagValues(cmd, "charge-type", "Month", "Dynamic", "Year")
	command.SetFlagValues(cmd, "mode", "Normal", "HA")

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, common.GetProjectId(), common.GetRegion())
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, vpcID, common.GetProjectId(), common.GetRegion())
	})
	command.SetCompletion(cmd, "param-group-id", func() []string {
		return listParamTemplates(ctx, version, common.GetProjectId(), common.GetRegion(), common.GetZone())
	})
	command.SetFlagValues(cmd, "machine-type", dbMachineTypeList...)

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("machine-type")

	// 自定义 usage，突出必填参数
	requiredFlags := []string{"name", "password", "version", "machine-type"}
	cmd.SetUsageFunc(func(c *cobra.Command) error {
		w := c.OutOrStderr()
		fmt.Fprintln(w, "Usage:")
		fmt.Fprintf(w, "  %s [flags]\n\n", c.CommandPath())
		fmt.Fprintln(w, "★ Required flags (must be provided):")
		for _, name := range requiredFlags {
			f := c.Flags().Lookup(name)
			if f != nil {
				fmt.Fprintf(w, "  --%-20s %s\n", f.Name, f.Usage)
			}
		}
		fmt.Fprintln(w, "\nOptional flags:")
		c.Flags().VisitAll(func(f *pflag.Flag) {
			for _, req := range requiredFlags {
				if f.Name == req {
					return
				}
			}
			defVal := ""
			if f.DefValue != "" && f.DefValue != "[]" {
				defVal = fmt.Sprintf(" (default %s)", f.DefValue)
			}
			fmt.Fprintf(w, "  --%-20s %s%s\n", f.Name, f.Usage, defVal)
		})
		return nil
	})

	return cmd
}

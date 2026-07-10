package sqlserver

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

var dbStorageClassList = []string{"CLOUD_RSSD"}
var dbSpecClassList = []string{"O"}
var dbMachineTypeList = []string{
	"o.sqlserver2m.medium",  // 2C4G
	"o.sqlserver2m.xlarge",  // 4C8G
	"o.sqlserver2m.2xlarge", // 8C16G
	"o.sqlserver2m.4xlarge", // 16C32G
	"o.sqlserver2m.8xlarge", // 32C64G
	"o.sqlserver4m.medium",  // 2C8G
	"o.sqlserver4m.xlarge",  // 4C16G
	"o.sqlserver4m.2xlarge", // 8C32G
	"o.sqlserver4m.4xlarge", // 16C64G
	"o.sqlserver4m.8xlarge", // 32C128G
	"o.sqlserver8m.medium",  // 2C16G
	"o.sqlserver8m.xlarge",  // 4C32G
	"o.sqlserver8m.2xlarge", // 8C64G
	"o.sqlserver8m.4xlarge", // 16C128G
	"o.sqlserver8m.8xlarge", // 32C256G
}

// newCreate returns the "create" command for SQL Server Normal (single-node) instances.
func newCreate(ctx *cli.Context) *cobra.Command {
	var labels []string
	var name, password, version, storageClass, specClass string
	var cpu, memory, port, diskSpace int
	var chargeType string
	var quantity int
	var vpcID, subnetID string
	var backupCount, backupTime, backupDuration int
	var tag, alarmTemplateID string
	var couponID string
	var async bool
	var common request.CommonBase

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create SQL Server instance (Normal/single-node mode) on UCloud platform",
		Long:  "Create SQL Server instance (Normal/single-node mode) on UCloud platform",
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

			params := map[string]interface{}{
				"Action":             "CreateUDBSQLServerInstance",
				"Region":             region,
				"Zone":               zone,
				"Name":               name,
				"AdminPassword":      password,
				"DBTypeId":           version,
				"Port":               port,
				"DiskSpace":          diskSpace,
				"CPU":                cpu,
				"MemoryLimit":        memory,
				"StorageClass":       storageClass,
				"SpecificationClass": specClass,
				"ChargeType":         chargeType,
				"Quantity":           quantity,
				"InstanceMode":       "Normal",
				"BackupCount":        backupCount,
				"BackupTime":         backupTime,
				"BackupDuration":     backupDuration,
			}
			if projectID != "" {
				params["ProjectId"] = projectID
			}

			// Optional params, only send when explicitly set
			if c.Flags().Changed("vpc-id") {
				params["VPCId"] = vpcID
			}
			if c.Flags().Changed("subnet-id") {
				params["SubnetId"] = subnetID
			}
			if c.Flags().Changed("tag") {
				params["Tag"] = tag
			}
			if c.Flags().Changed("alarm-template-id") {
				params["AlarmTemplateId"] = alarmTemplateID
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
	flags.StringVar(&version, "db-type", "", "Required. SQL Server version. Options: sqlserver-2017, sqlserver-2019, sqlserver-2022")

	// Optional flags with defaults
	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)
	ctx.BindProjectID(cmd, &common)
	flags.IntVar(&cpu, "cpu", 2, "Optional. CPU cores. Options: 2/4/8/16/32/64, default 2")
	flags.IntVar(&memory, "memory", 4000, "Optional. Memory limit (MB). Options: 2000/4000/6000/8000/12000/16000/24000/32000/48000/64000/96000/128000/192000/256000/320000, default 4000")
	flags.IntVar(&port, "port", 1433, "Optional. Port, default 1433")
	flags.IntVar(&diskSpace, "disk-size-gb", 50, "Optional. Disk size (GiB), 20-32000, default 50")
	flags.StringVar(&storageClass, "storage-class", "CLOUD_RSSD", "Optional. Storage class: CLOUD_RSSD")
	flags.StringVar(&specClass, "spec-class", "O", "Optional. Spec class: O(NVMe)")

	flags.StringVar(&chargeType, "charge-type", "Month", "Optional. Year / Month / Dynamic")
	flags.IntVar(&quantity, "quantity", 1, "Optional. Purchase duration")
	flags.StringVar(&vpcID, "vpc-id", "", "Optional. VPC ID. See 'ucloud vpc list'")
	flags.StringVar(&subnetID, "subnet-id", "", "Optional. Subnet ID. See 'ucloud subnet list'")
	flags.IntVar(&backupCount, "backup-count", 7, "Optional. Weekly backup count, default 7")
	flags.IntVar(&backupTime, "backup-time", 1, "Optional. Backup start hour (0-23), default 1")
	flags.IntVar(&backupDuration, "backup-duration", 24, "Optional. Backup interval hours, default 24")
	flags.StringVar(&tag, "tag", "", "Optional. Business group name")
	flags.StringVar(&alarmTemplateID, "alarm-template-id", "", "Optional. Alarm template ID")
	flags.StringSliceVar(&labels, "label", nil, "Optional. Resource label, format: key=value, repeatable")
	flags.StringVar(&couponID, "coupon-id", "", "Optional. Coupon ID")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for creation to finish")

	command.SetFlagValues(cmd, "db-type", dbVersionList...)
	command.SetFlagValues(cmd, "storage-class", dbStorageClassList...)
	command.SetFlagValues(cmd, "spec-class", dbSpecClassList...)
	command.SetFlagValues(cmd, "charge-type", "Month", "Dynamic", "Year")
	command.SetFlagValues(cmd, "cpu", "2", "4", "8", "16", "32", "64")

	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, common.GetProjectId(), common.GetRegion())
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, vpcID, common.GetProjectId(), common.GetRegion())
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("db-type")

	// Custom usage, highlight required flags
	requiredFlags := []string{"name", "password", "db-type"}
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

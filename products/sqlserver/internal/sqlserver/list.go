package sqlserver

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList returns the "list" command for SQL Server instances.
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SQL Server instances",
		Long:  "List SQL Server instances",
		Run: func(c *cobra.Command, args []string) {
			if *req.DBId != "" {
				*req.DBId = ctx.PickResourceID(*req.DBId)
			}
			resp, err := client.DescribeUDBInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []UDBSQLServerRow{}
			for _, ins := range resp.DataSet {
				row := UDBSQLServerRow{}
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
					row := UDBSQLServerRow{}
					row.Name = slave.Name
					row.Zone = slave.Zone
					row.Role = fmt.Sprintf("⮑ %s", slave.Role)
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
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBId = flags.String("udb-id", "", "Optional. List the specified SQL Server instance")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindLimit(cmd, req)
	ctx.BindOffset(cmd, req)
	req.IncludeSlaves = flags.Bool("include-slaves", false, "Optional. When specifying the udb-id, whether to display its slaves together. Accept values:true, false")
	req.ClassType = sdk.String("sqlserver")

	command.SetFlagValues(cmd, "include-slaves", "true", "false")
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sqlserver", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

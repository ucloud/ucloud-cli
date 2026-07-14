package pgsql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newGet ucloud pgsql db get
func newGet(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewGetUPgSQLInstanceRequest()
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display details of a UPgSQL instance",
		Long:  "Display details of a UPgSQL instance",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			resp, err := client.GetUPgSQLInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ins := resp.DataSet
			if ins.InstanceID == "" {
				ctx.HandleError(fmt.Errorf("pgsql[%s] may not exist", *req.InstanceID))
				return
			}
			attrs := []cli.DescribeRow{
				{Attribute: "InstanceID", Content: ins.InstanceID},
				{Attribute: "Name", Content: ins.Name},
				{Attribute: "State", Content: ins.State},
				{Attribute: "Zone", Content: ins.Zone},
				{Attribute: "BackupZone", Content: ins.BackupZone},
				{Attribute: "DBVersion", Content: ins.DBVersion},
				{Attribute: "InstanceMode", Content: ins.InstanceMode},
				{Attribute: "AdminUser", Content: ins.AdminUser},
				{Attribute: "IP", Content: ins.IP},
				{Attribute: "Port", Content: strconv.Itoa(ins.Port)},
				{Attribute: "VPCID", Content: ins.VPCID},
				{Attribute: "SubnetID", Content: ins.SubnetID},
				{Attribute: "ParamGroupID", Content: strconv.Itoa(ins.ParamGroupID)},
				{Attribute: "MemoryLimit", Content: fmt.Sprintf("%dMB", ins.MemoryLimit)},
				{Attribute: "DiskSpace", Content: fmt.Sprintf("%dGB", ins.DiskSpace)},
				{Attribute: "DiskUsedSize", Content: fmt.Sprintf("%.2fGB", ins.DiskUsedSize)},
				{Attribute: "Remark", Content: ins.Remark},
				{Attribute: "BackupCount", Content: strconv.Itoa(ins.BackupCount)},
				{Attribute: "BackupBeginTime", Content: strconv.Itoa(ins.BackupBeginTime)},
				{Attribute: "BackupDate", Content: ins.BackupDate},
				{Attribute: "CreateTime", Content: common.FormatDateTime(ins.CreateTime)},
				{Attribute: "ModifyTime", Content: common.FormatDateTime(ins.ModifyTime)},
				{Attribute: "ExpiredTime", Content: common.FormatDateTime(ins.ExpiredTime)},
			}
			fmt.Fprintln(ctx.ProgressWriter(), "Attributes:")
			ctx.PrintList(attrs)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("instance-id")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}

package tidb

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDescribe ucloud utidb describe
func newDescribe(ctx *cli.Context) *cobra.Command {
	var id string
	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewGetTiDBClusterServiceRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of a UTiDB instance",
		Long:  "Show details of a UTiDB instance",
		Run: func(c *cobra.Command, args []string) {
			req.Id = sdk.String(ctx.PickResourceID(id))
			resp, err := client.GetTiDBClusterService(req)
			if err != nil {
				handleAPIError(ctx, err)
				return
			}
			d := resp.Data
			rows := []cli.DescribeRow{
				{Attribute: "ID", Content: d.Id},
				{Attribute: "Name", Content: d.Name},
				{Attribute: "State", Content: d.State},
				{Attribute: "Port", Content: fmt.Sprintf("%d", d.Port)},
				{Attribute: "IP", Content: d.Ip},
				{Attribute: "VPCId", Content: d.VPCId},
				{Attribute: "SubnetId", Content: d.SubnetId},
				{Attribute: "Version", Content: d.Version},
				{Attribute: "CreateTime", Content: fmt.Sprintf("%d", d.CreateTime)},
				{Attribute: "DTType", Content: fmt.Sprintf("%d", d.DTType)},
				{Attribute: "AutoBackup", Content: d.AutoBackup},
				{Attribute: "BinlogState", Content: d.BinlogState},
				{Attribute: "TiFlashState", Content: d.TiFlashState},
				{Attribute: "DashboardUrl", Content: d.DashboardUrl},
				{Attribute: "GrafanaUrl", Content: d.GrafanaUrl},
			}
			ctx.PrintList(rows)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&id, "utidb-id", "", "Required. Resource ID of the UTiDB instance")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("utidb-id")
	command.SetCompletion(cmd, "utidb-id", func() []string {
		return listResourceIDs(ctx, nil, req.GetRegion(), req.GetZone(), req.GetProjectId())
	})

	return cmd
}

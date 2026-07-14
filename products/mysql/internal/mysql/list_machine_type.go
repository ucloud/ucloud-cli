package mysql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListMachineType ucloud mysql db list-machine-type
func newListMachineType(ctx *cli.Context) *cobra.Command {
	var mode string
	var common request.CommonBase
	cmd := &cobra.Command{
		Use:   "list-machine-type",
		Short: "List available MySQL machine types",
		Long:  "List available MySQL machine types via ListUDBMachineType API",
		Run: func(c *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, udb.NewClient)
			req := client.NewListUDBMachineTypeRequest()
			req.Region = sdk.String(common.GetRegion())
			req.Zone = sdk.String(common.GetZone())
			req.ProjectId = sdk.String(common.GetProjectId())
			if mode != "" {
				req.InstanceMode = &mode
			}
			resp, err := client.ListUDBMachineType(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			var rows []MachineTypeRow
			for _, mt := range resp.DataSet {
				rows = append(rows, MachineTypeRow{
					ID:          mt.ID,
					Description: mt.Description,
					Cpu:         mt.Cpu,
					Memory:      mt.Memory,
					Group:       mt.Group,
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)
	ctx.BindProjectID(cmd, &common)
	flags.StringVar(&mode, "mode", "", "Optional. Instance mode: Normal / HA")
	command.SetFlagValues(cmd, "mode", "Normal", "HA")

	return cmd
}

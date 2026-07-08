package mysql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// UDBConfRow 表格行
type UDBConfRow struct {
	ConfID      int
	DBVersion   string
	Name        string
	Description string
	Modifiable  bool
	Zone        string
}

// newUDBConfList ucloud mysql conf list
func newUDBConfList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configuartion files of MySQL instances",
		Long:  "List configuartion files of MySQL instances",
		Run: func(c *cobra.Command, args []string) {
			if *req.GroupId == 0 {
				req.GroupId = nil
			}
			resp, err := client.DescribeUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []UDBConfRow{}
			for _, ins := range resp.DataSet {
				row := UDBConfRow{
					ConfID:      ins.GroupId,
					Name:        ins.GroupName,
					Zone:        ins.Zone,
					DBVersion:   ins.DBTypeId,
					Description: ins.Description,
					Modifiable:  ins.Modifiable,
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindOffset(cmd, req)
	ctx.BindLimit(cmd, req)
	req.GroupId = flags.Int("conf-id", 0, "Optional. Configuration identifier for the configuration to be described")
	req.ClassType = sdk.String("sql")

	command.SetCompletion(cmd, "conf-id", func() []string {
		return getConfIDList(ctx, *req.ClassType, *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

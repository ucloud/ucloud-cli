package mysql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// UDBConfParamRow 参数配置展示表格行
type UDBConfParamRow struct {
	Key   string
	Value string
}

// newUDBConfDescribe ucloud udb conf describe
func newUDBConfDescribe(ctx *cli.Context) *cobra.Command {
	var confID string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBParamGroupRequest()
	req.RegionFlag = sdk.Bool(false)
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Display details about a configuration file of MySQL instance",
		Long:  "Display details about a configuration file of MySQL instance",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.GroupId = &id
			resp, err := client.DescribeUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.DataSet) != 1 {
				ctx.HandleError(fmt.Errorf("conf-id[%d] may not be exist", *req.GroupId))
				return
			}
			conf := resp.DataSet[0]
			attrs := []cli.DescribeRow{
				{Attribute: "ConfID", Content: strconv.Itoa(conf.GroupId)},
				{Attribute: "DBVersion", Content: conf.DBTypeId},
				{Attribute: "Name", Content: conf.GroupName},
				{Attribute: "Description", Content: conf.Description},
				{Attribute: "Modifiable", Content: strconv.FormatBool(conf.Modifiable)},
				{Attribute: "Zone", Content: conf.Zone},
			}
			fmt.Fprintln(ctx.ProgressWriter(), "Attributes:")
			ctx.PrintList(attrs)

			params := []UDBConfParamRow{}
			for _, p := range conf.ParamMember {
				if p.Value == "" {
					continue
				}
				row := UDBConfParamRow{
					Key:   p.Key,
					Value: p.Value,
				}
				params = append(params, row)
			}
			fmt.Fprintln(ctx.ProgressWriter(), "\nParameters:")
			ctx.PrintList(params)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Requried. Configuration identifier for the configuration to be described")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-id")
	command.SetCompletion(cmd, "conf-id", func() []string {
		return getConfIDList(ctx, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

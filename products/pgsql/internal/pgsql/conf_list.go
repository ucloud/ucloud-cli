package pgsql

import (
	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newConfList ucloud pgsql conf list
func newConfList(ctx *cli.Context) *cobra.Command {
	var common request.CommonBase
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UPgSQL parameter templates",
		Long:  "List UPgSQL parameter templates",
		Run: func(c *cobra.Command, args []string) {
			templates, err := listParamTemplates(ctx, common.GetProjectId(), common.GetRegion(), common.GetZone())
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []PgsqlConfRow{}
			for _, t := range templates {
				rows = append(rows, PgsqlConfRow{
					GroupID:     t.GroupID,
					GroupName:   t.GroupName,
					DBVersion:   t.DBVersion,
					Description: t.Description,
					Modifiable:  t.Modifiable,
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindProjectID(cmd, &common)
	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)

	return cmd
}

package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newListMachineType ucloud pgsql db list-machine-type
func newListMachineType(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewListUPgSQLMachineTypeRequest()
	cmd := &cobra.Command{
		Use:   "list-machine-type",
		Short: "List available UPgSQL machine types",
		Long:  "List available UPgSQL machine types via ListUPgSQLMachineType API",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.ListUPgSQLMachineType(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []PgsqlMachineTypeRow{}
			for _, mt := range resp.DataSet {
				rows = append(rows, PgsqlMachineTypeRow{
					ID:          mt.ID,
					Description: mt.Description,
					Cpu:         mt.Cpu,
					Memory:      mt.Memory,
					Os:          mt.Os,
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	return cmd
}

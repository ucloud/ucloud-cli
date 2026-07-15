package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newListVersion ucloud pgsql db list-version
func newListVersion(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewListUPgSQLVersionRequest()
	cmd := &cobra.Command{
		Use:   "list-version",
		Short: "List available UPgSQL versions",
		Long:  "List available UPgSQL versions via ListUPgSQLVersion API",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.ListUPgSQLVersion(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []PgsqlVersionRow{}
			for _, v := range resp.DataSet {
				rows = append(rows, PgsqlVersionRow{
					DBVersion: v.DBVersion,
					Available: v.Available,
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

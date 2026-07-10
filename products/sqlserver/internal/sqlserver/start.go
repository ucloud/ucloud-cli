package sqlserver

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStart returns the "start" command for SQL Server instances.
func newStart(ctx *cli.Context) *cobra.Command {
	var async bool
	var idNames []string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewStartUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start SQL Server instances by udb-id",
		Long:  "Start SQL Server instances by udb-id",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.DBId = &id
				_, err := client.StartUDBInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(w, "udb[%s] is starting\n", idname)
				} else {
					text := fmt.Sprintf("udb[%s] is starting", idname)
					ctx.PollerTo(w, describeUdbByID(ctx)).Spoll(*req.DBId, text, []string{UDB_RUNNING, UDB_FAIL})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "start", Status: "Starting"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to start")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	cmd.MarkFlagRequired("udb-id")

	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, []string{UDB_SHUTOFF}, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

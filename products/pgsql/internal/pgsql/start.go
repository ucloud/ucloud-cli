package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStart ucloud pgsql db start
func newStart(ctx *cli.Context) *cobra.Command {
	var async bool
	var idNames []string
	client := newUPgSQLClient(ctx)
	req := client.NewStartUPgSQLInstanceRequest()
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start UPgSQL instances by instance-id",
		Long:  "Start UPgSQL instances by instance-id",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.InstanceID = sdk.String(id)
				_, err := client.StartUPgSQLInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(w, "pgsql[%s] is starting\n", idname)
				} else {
					text := fmt.Sprintf("pgsql[%s] is starting", idname)
					ctx.PollerTo(w, describePgsqlByID(ctx)).Spoll(id, text, []string{PGSQL_RUNNING, PGSQL_START_FAILED})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "start", Status: "Starting"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "instance-id", nil, "Required. Resource ID of UPgSQL instances to start")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	cmd.MarkFlagRequired("instance-id")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})
	return cmd
}

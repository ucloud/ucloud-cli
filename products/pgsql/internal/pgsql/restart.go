package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRestart ucloud pgsql db restart
func newRestart(ctx *cli.Context) *cobra.Command {
	var async bool
	var idNames []string
	client := newUPgSQLClient(ctx)
	req := client.NewRestartUPgSQLInstanceRequest()
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart UPgSQL instances by instance-id",
		Long:  "Restart UPgSQL instances by instance-id",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.InstanceID = sdk.String(id)
				_, err := client.RestartUPgSQLInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(w, "pgsql[%s] is restarting\n", idname)
				} else {
					text := fmt.Sprintf("pgsql[%s] is restarting", idname)
					ctx.PollerTo(w, describePgsqlByID(ctx)).Spoll(id, text, []string{PGSQL_RUNNING, PGSQL_START_FAILED, PGSQL_SHUTDOWN_FAILED})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "restart", Status: "Restarting"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "instance-id", nil, "Required. Resource ID of UPgSQL instances to restart")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	req.ForceToRestart = flags.Bool("force", false, "Optional. Restart UPgSQL instances by force or not")
	req.RestartHost = flags.Bool("restart-host", false, "Optional. Restart the host together or not")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	cmd.MarkFlagRequired("instance-id")

	command.SetFlagValues(cmd, "force", "true", "false")
	command.SetFlagValues(cmd, "restart-host", "true", "false")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})
	return cmd
}

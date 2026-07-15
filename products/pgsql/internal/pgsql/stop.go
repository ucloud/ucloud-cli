package pgsql

import (
	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStop ucloud pgsql db stop
func newStop(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var async bool
	client := newUPgSQLClient(ctx)
	req := client.NewStopUPgSQLInstanceRequest()
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop UPgSQL instances by instance-id",
		Long:  "Stop UPgSQL instances by instance-id",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.InstanceID = sdk.String(id)
				if err := stopPgsqlIns(ctx, req, async, w); err != nil {
					ctx.HandleError(err)
					continue
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "stop", Status: "Stopping"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "instance-id", nil, "Required. Resource ID of UPgSQL instances to stop")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	req.ForceToStop = flags.Bool("force", false, "Optional. Stop UPgSQL instances by force or not")
	req.StopHost = flags.Bool("stop-host", false, "Optional. Stop the host together or not")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	cmd.MarkFlagRequired("instance-id")

	command.SetFlagValues(cmd, "force", "true", "false")
	command.SetFlagValues(cmd, "stop-host", "true", "false")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}

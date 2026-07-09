package mysql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStop ucloud udb stop
func newStop(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var async bool
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewStopUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop MySQL instances by udb-id",
		Long:  "Stop MySQL instances by udb-id",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.DBId = sdk.String(id)
				if err := stopUdbIns(ctx, req, async, w); err != nil {
					continue
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "stop", Status: "Stopping"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to stop")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.ForceToKill = flags.Bool("force", false, "Optional. Stop UDB instances by force or not")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")

	cmd.MarkFlagRequired("udb-id")

	command.SetFlagValues(cmd, "force", "true", "false")
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, []string{UDB_RUNNING}, "", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

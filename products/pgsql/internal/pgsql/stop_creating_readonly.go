package pgsql

import (
	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStopCreatingReadonly ucloud pgsql db stop-creating-readonly
func newStopCreatingReadonly(ctx *cli.Context) *cobra.Command {
	var idNames []string
	client := newUPgSQLClient(ctx)
	req := client.NewStopUPgSQLCreatingReadonlyRequest()
	cmd := &cobra.Command{
		Use:   "stop-creating-readonly",
		Short: "Stop readonly replicas that are still being created",
		Long:  "Stop readonly replicas of a UPgSQL instance that are still being created",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.InstanceID = sdk.String(id)
				_, err := client.StopUPgSQLCreatingReadonly(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "stop-creating-readonly", Status: "Stopped"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "instance-id", nil, "Required. Resource ID of the readonly replicas to stop creating")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("instance-id")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}

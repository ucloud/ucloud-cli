package umongodb

import (
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-sdk-go/services/umongodb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRestart implements `umongodb restart`.
func newRestart(ctx *cli.Context) *cobra.Command {
	var async bool
	var ids []string

	client := cli.NewServiceClient(ctx, umongodb.NewClient)
	req := client.NewRestartUMongoDBClusterRequest()

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart MongoDB instances",
		Long:  "Restart one or more MongoDB instances.",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)
				req.ClusterId = sdk.String(id)
				_, err := client.RestartUMongoDBCluster(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				text := fmt.Sprintf("%s[%s] is restarting", productName, id)
				if async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeByID(ctx, *req.Region, *req.Zone)).Spoll(id, text, []string{stateRunning, stateFail})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "restart", Status: "Restarting"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. Cluster ID(s) of MongoDB instances to restart.")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the operation to finish.")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getMongoDBIDList(ctx, nil, *req.Region, *req.Zone, *req.ProjectId)
	})

	return cmd
}

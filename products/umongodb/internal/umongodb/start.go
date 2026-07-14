package umongodb

import (
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-sdk-go/services/umongodb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newStart implements `umongodb start`.
func newStart(ctx *cli.Context) *cobra.Command {
	var async bool
	var ids []string

	client := cli.NewServiceClient(ctx, umongodb.NewClient)
	req := client.NewStartUMongoDBClusterRequest()

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start MongoDB instances",
		Long:  "Start one or more stopped MongoDB instances.",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)
				req.ClusterId = sdk.String(id)
				_, err := client.StartUMongoDBCluster(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				text := fmt.Sprintf("%s[%s] is starting", productName, id)
				if async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeByID(ctx, *req.Region, *req.Zone)).Spoll(id, text, []string{stateRunning, stateFail})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "start", Status: "Starting"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. Cluster ID(s) of MongoDB instances to start.")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the operation to finish.")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getMongoDBIDList(ctx, []string{stateStopped}, *req.Region, *req.Zone, *req.ProjectId)
	})

	return cmd
}

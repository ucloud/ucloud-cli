package clickhouse

import (
	"fmt"

	"github.com/spf13/cobra"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRestart ucloud clickhouse restart
func newRestart(ctx *cli.Context) *cobra.Command {
	var async *bool
	var clusterID *string
	var yes *bool
	client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
	req := client.NewRestartUClickhouseClusterServiceRequest()
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart UClickhouse cluster service",
		Long:  "Restart UClickhouse cluster service",
		Args:  noArgs,
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*clusterID)
			ok, err := ctx.Confirm(*yes, fmt.Sprintf("Are you sure to restart UClickhouse cluster[%s]?", id))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			req.ClusterId = &id
			w := ctx.ProgressWriter()
			_, err = restartUClickhouseClusterService(client, req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("clickhouse[%s] is restarting", id)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeClusterByID(ctx)).Spoll(id, text, []string{STATUS_RUNNING, STATUS_RESTART_FAILED})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "restart", Status: "Restarting"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	clusterID = flags.String("clickhouse-id", "", "Required. UClickhouse cluster ID to restart")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	yes = flags.BoolP("yes", "y", false, "Optional. Skip confirmation prompt")
	async = flags.Bool("async", false, "Optional. Do not wait for restart to finish")

	command.SetCompletion(cmd, "clickhouse-id", func() []string {
		return getClusterList(ctx, []string{STATUS_RUNNING, STATUS_RESTART_FAILED}, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("clickhouse-id")
	return cmd
}

func restartUClickhouseClusterService(client *uclickhousesdk.UClickhouseClient, req *uclickhousesdk.RestartUClickhouseClusterServiceRequest) (*opResponse, error) {
	var resp opResponse
	reqCopier := *req
	err := invokeUClickhouseAction(client, "RestartUClickhouseClusterService", &reqCopier, &resp)
	return &resp, err
}

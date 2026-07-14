package clickhouse

import (
	"fmt"

	"github.com/spf13/cobra"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newExpand ucloud clickhouse expand
func newExpand(ctx *cli.Context) *cobra.Command {
	var async *bool
	var clusterID *string
	client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
	req := client.NewExpandUClickhouseClusterRequest()
	cmd := &cobra.Command{
		Use:   "expand",
		Short: "Expand UClickhouse cluster node count",
		Long:  "Expand UClickhouse cluster node count",
		Args:  noArgs,
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*clusterID)
			req.ClusterId = &id
			w := ctx.ProgressWriter()
			_, err := expandUClickhouseCluster(client, req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("clickhouse[%s] is expanding", id)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeClusterByID(ctx)).Spoll(id, text, []string{STATUS_RUNNING, STATUS_EXPAND_FAILED})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "expand", Status: "Expanding"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	clusterID = flags.String("clickhouse-id", "", "Required. UClickhouse cluster ID to expand")
	req.TotalNodeCount = flags.Int("total-node-count", 0, "Required. Total node count after expansion")
	req.SyncNodeId = flags.String("sync-node-id", "", "Optional. Existing node ID used to sync schema/user information")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	async = flags.Bool("async", false, "Optional. Do not wait for expansion to finish")

	command.SetCompletion(cmd, "clickhouse-id", func() []string {
		return getClusterList(ctx, []string{STATUS_RUNNING}, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("clickhouse-id")
	cmd.MarkFlagRequired("total-node-count")
	return cmd
}

func expandUClickhouseCluster(client *uclickhousesdk.UClickhouseClient, req *uclickhousesdk.ExpandUClickhouseClusterRequest) (*opResponse, error) {
	var resp opResponse
	reqCopier := *req
	err := invokeUClickhouseAction(client, "ExpandUClickhouseCluster", &reqCopier, &resp)
	return &resp, err
}

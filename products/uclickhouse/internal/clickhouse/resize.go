package clickhouse

import (
	"fmt"

	"github.com/spf13/cobra"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newResize ucloud clickhouse resize
func newResize(ctx *cli.Context) *cobra.Command {
	var async *bool
	var clusterID *string
	client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
	req := client.NewResizeUClickhouseClusterRequest()
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize UClickhouse cluster",
		Long:  "Resize UClickhouse cluster. Set target-machine-type-id to change spec, or target-data-disk-size-gb to expand disk.",
		Args:  noArgs,
		Run: func(cmd *cobra.Command, args []string) {
			id := ctx.PickResourceID(*clusterID)
			req.ClusterId = &id
			w := ctx.ProgressWriter()
			_, err := resizeUClickhouseCluster(client, req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("clickhouse[%s] is resizing", id)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeClusterByID(ctx)).Spoll(id, text, []string{STATUS_RUNNING, STATUS_RESIZE_FAILED})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "resize", Status: "Resizing"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	clusterID = flags.String("clickhouse-id", "", "Required. UClickhouse cluster ID to resize")
	req.TargetMachineTypeId = flags.String("target-machine-type-id", "", "Optional. Target machine type ID")
	req.TargetDataDiskSize = flags.Int("target-data-disk-size-gb", 0, "Optional. Target data disk size in GB")
	req.IsZooKeeperNode = flags.Bool("zookeeper-node", false, "Optional. Resize Zookeeper nodes instead of ClickHouse nodes")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	async = flags.Bool("async", false, "Optional. Do not wait for resize to finish")

	command.SetCompletion(cmd, "clickhouse-id", func() []string {
		return getClusterList(ctx, []string{STATUS_RUNNING}, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("clickhouse-id")
	return cmd
}

func resizeUClickhouseCluster(client *uclickhousesdk.UClickhouseClient, req *uclickhousesdk.ResizeUClickhouseClusterRequest) (*opResponse, error) {
	var resp opResponse
	reqCopier := *req
	err := invokeUClickhouseAction(client, "ResizeUClickhouseCluster", &reqCopier, &resp)
	return &resp, err
}

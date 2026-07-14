package clickhouse

import (
	"fmt"

	"github.com/spf13/cobra"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud clickhouse delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var clusterIDs *[]string
	var yes *bool
	client := cli.NewServiceClient(ctx, uclickhousesdk.NewClient)
	req := client.NewDestroyUClickhouseClusterRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete UClickhouse clusters",
		Long:  "Delete UClickhouse clusters",
		Args:  noArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(*yes, "Are you sure to delete UClickhouse cluster(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idName := range *clusterIDs {
				id := ctx.PickResourceID(idName)
				req.ClusterId = &id
				_, err := destroyUClickhouseCluster(client, req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "clickhouse[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	clusterIDs = flags.StringSlice("clickhouse-id", nil, "Required. UClickhouse cluster ID(s) to delete")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	yes = flags.BoolP("yes", "y", false, "Optional. Skip confirmation prompt")

	command.SetCompletion(cmd, "clickhouse-id", func() []string {
		return getClusterList(ctx, []string{STATUS_RUNNING, STATUS_CREATE_FAILED, STATUS_RESTART_FAILED, STATUS_RESIZE_FAILED, STATUS_EXPAND_FAILED, STATUS_BACKUP_RESTORE_FAILED}, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("clickhouse-id")
	return cmd
}

func destroyUClickhouseCluster(client *uclickhousesdk.UClickhouseClient, req *uclickhousesdk.DestroyUClickhouseClusterRequest) (*opResponse, error) {
	var resp opResponse
	reqCopier := *req
	err := invokeUClickhouseAction(client, "DestroyUClickhouseCluster", &reqCopier, &resp)
	return &resp, err
}

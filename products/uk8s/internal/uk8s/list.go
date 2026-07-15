package uk8s

import (
	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewListUK8SClusterV2Request()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UK8S clusters",
		Long:  "List UK8S clusters in the active region and project.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if req.ClusterId != nil && *req.ClusterId != "" {
				*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			}
			resp, err := client.ListUK8SClusterV2(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if ctx.Format() != cli.OutputTable {
				ctx.PrintList(resp)
				return
			}
			ctx.PrintList(clusterRows(resp.ClusterSet))
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterId = flags.String("cluster-id", "", "Optional. List only the specified cluster.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindLimit(cmd, req)
	ctx.BindOffset(cmd, req)
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, nil, derefStr(req.Region), derefStr(req.ProjectId))
	})
	return cmd
}

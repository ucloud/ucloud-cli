package uk8s

import (
	"fmt"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewDescribeUK8SClusterRequest()

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of a UK8S cluster",
		Long:  "Show the attributes of one UK8S cluster.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			cluster, err := client.DescribeUK8SCluster(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if cluster.ClusterId == "" {
				ctx.HandleError(fmt.Errorf("cluster %q not found", *req.ClusterId))
				return
			}
			if ctx.Format() != cli.OutputTable {
				ctx.PrintList(cluster)
				return
			}

			ctx.PrintList(clusterDescribeRows(cluster))
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterId = flags.String("cluster-id", "", "Required. Cluster ID to describe.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("cluster-id")
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, nil, derefStr(req.Region), derefStr(req.ProjectId))
	})
	return cmd
}

package uk8s

import (
	"fmt"
	"time"

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

			created := ""
			if cluster.CreateTime > 0 {
				created = time.Unix(int64(cluster.CreateTime), 0).Format(time.RFC3339)
			}
			rows := []cli.DescribeRow{
				{Attribute: "ResourceID", Content: cluster.ClusterId},
				{Attribute: "Name", Content: cluster.ClusterName},
				{Attribute: "Version", Content: cluster.Version},
				{Attribute: "Status", Content: cluster.Status},
				{Attribute: "VPCID", Content: cluster.VPCId},
				{Attribute: "SubnetID", Content: cluster.SubnetId},
				{Attribute: "ServiceCIDR", Content: cluster.ServiceCIDR},
				{Attribute: "PodCIDR", Content: cluster.PodCIDR},
				{Attribute: "ClusterDomain", Content: cluster.ClusterDomain},
				{Attribute: "MasterCount", Content: fmt.Sprintf("%d", cluster.MasterCount)},
				{Attribute: "NodeCount", Content: fmt.Sprintf("%d", cluster.NodeCount)},
				{Attribute: "APIServer", Content: cluster.ApiServer},
				{Attribute: "ExternalAPIServer", Content: cluster.ExternalApiServer},
				{Attribute: "KubeProxyMode", Content: cluster.KubeProxy.Mode},
				{Attribute: "Created", Content: created},
			}
			ctx.PrintList(rows)
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

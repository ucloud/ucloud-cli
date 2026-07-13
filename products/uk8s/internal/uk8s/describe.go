package uk8s

import (
	"encoding/json"
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
			if ctx.Format() != cli.OutputTable {
				ctx.PrintList(cluster)
				return
			}

			created := ""
			if cluster.CreateTime > 0 {
				created = time.Unix(int64(cluster.CreateTime), 0).Format(time.RFC3339)
			}
			rows := []cli.DescribeRow{
				{Attribute: "ResourceID", Content: cluster.ClusterId},
				{Attribute: "Name", Content: cluster.ClusterName},
				{Attribute: "ClusterId", Content: cluster.ClusterId},
				{Attribute: "ClusterName", Content: cluster.ClusterName},
				{Attribute: "ApiServer", Content: cluster.ApiServer},
				{Attribute: "Version", Content: cluster.Version},
				{Attribute: "CreateTime", Content: fmt.Sprintf("%d", cluster.CreateTime)},
				{Attribute: "Status", Content: cluster.Status},
				{Attribute: "VPCID", Content: cluster.VPCId},
				{Attribute: "VPCId", Content: cluster.VPCId},
				{Attribute: "SubnetID", Content: cluster.SubnetId},
				{Attribute: "SubnetId", Content: cluster.SubnetId},
				{Attribute: "ServiceCIDR", Content: cluster.ServiceCIDR},
				{Attribute: "PodCIDR", Content: cluster.PodCIDR},
				{Attribute: "ClusterDomain", Content: cluster.ClusterDomain},
				{Attribute: "MasterCount", Content: fmt.Sprintf("%d", cluster.MasterCount)},
				{Attribute: "NodeCount", Content: fmt.Sprintf("%d", cluster.NodeCount)},
				{Attribute: "MasterResourceStatus", Content: cluster.MasterResourceStatus},
				{Attribute: "APIServer", Content: cluster.ApiServer},
				{Attribute: "ExternalAPIServer", Content: cluster.ExternalApiServer},
				{Attribute: "ExternalApiServer", Content: cluster.ExternalApiServer},
				{Attribute: "KubeProxyMode", Content: cluster.KubeProxy.Mode},
				{Attribute: "KubeProxy", Content: fmt.Sprintf("%+v", cluster.KubeProxy)},
				{Attribute: "CACert", Content: cluster.CACert},
				{Attribute: "EtcdCert", Content: cluster.EtcdCert},
				{Attribute: "EtcdKey", Content: cluster.EtcdKey},
				{Attribute: "Created", Content: created},
			}
			rows = appendUhostInfoRows(rows, "MasterList", cluster.MasterList)
			rows = appendUhostInfoRows(rows, "NodeList", cluster.NodeList)
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

func appendUhostInfoRows(rows []cli.DescribeRow, group string, nodes []uk8ssdk.UhostInfo) []cli.DescribeRow {
	rows = append(rows, cli.DescribeRow{Attribute: "--- " + group + " ---", Content: fmt.Sprintf("%d nodes", len(nodes))})
	for i, node := range nodes {
		prefix := fmt.Sprintf("%s[%d]", group, i)
		ipSet, _ := json.Marshal(node.IPSet)
		diskSet, _ := json.Marshal(node.DiskSet)
		rows = append(rows,
			cli.DescribeRow{Attribute: prefix + ".NodeId", Content: node.NodeId},
			cli.DescribeRow{Attribute: prefix + ".Name", Content: node.Name},
			cli.DescribeRow{Attribute: prefix + ".NodeType", Content: node.NodeType},
			cli.DescribeRow{Attribute: prefix + ".Zone", Content: node.Zone},
			cli.DescribeRow{Attribute: prefix + ".State", Content: node.State},
			cli.DescribeRow{Attribute: prefix + ".CPU", Content: fmt.Sprintf("%d", node.CPU)},
			cli.DescribeRow{Attribute: prefix + ".Memory", Content: fmt.Sprintf("%d", node.Memory)},
			cli.DescribeRow{Attribute: prefix + ".OsName", Content: node.OsName},
			cli.DescribeRow{Attribute: prefix + ".CreateTime", Content: fmt.Sprintf("%d", node.CreateTime)},
			cli.DescribeRow{Attribute: prefix + ".ExpireTime", Content: fmt.Sprintf("%d", node.ExpireTime)},
			cli.DescribeRow{Attribute: prefix + ".IPSet", Content: string(ipSet)},
			cli.DescribeRow{Attribute: prefix + ".DiskSet", Content: string(diskSet)},
		)
	}
	return rows
}

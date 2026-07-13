package uk8s

import (
	"fmt"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newNodeList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewListUK8SClusterNodeV2Request()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List nodes in a UK8S cluster",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			resp, err := client.ListUK8SClusterNodeV2(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]cli.DescribeRow, 0, len(resp.NodeSet)*16)
			for i, node := range resp.NodeSet {
				rows = appendNodeInfoRows(rows, i, node)
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterId = flags.String("cluster-id", "", "Required. Cluster ID.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("cluster-id")
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, nil, derefStr(req.Region), derefStr(req.ProjectId))
	})
	return cmd
}

func appendNodeInfoRows(rows []cli.DescribeRow, index int, node uk8ssdk.NodeInfoV2) []cli.DescribeRow {
	prefix := fmt.Sprintf("Node[%d]", index)
	rows = append(rows,
		cli.DescribeRow{Attribute: "--- " + prefix + " ---", Content: node.NodeId},
		cli.DescribeRow{Attribute: prefix + ".AsgId", Content: node.AsgId},
		cli.DescribeRow{Attribute: prefix + ".CPU", Content: fmt.Sprintf("%d", node.CPU)},
		cli.DescribeRow{Attribute: prefix + ".CreateTime", Content: fmt.Sprintf("%d", node.CreateTime)},
		cli.DescribeRow{Attribute: prefix + ".ExpireTime", Content: fmt.Sprintf("%d", node.ExpireTime)},
		cli.DescribeRow{Attribute: prefix + ".GPU", Content: fmt.Sprintf("%d", node.GPU)},
		cli.DescribeRow{Attribute: prefix + ".InstanceId", Content: node.InstanceId},
		cli.DescribeRow{Attribute: prefix + ".InstanceName", Content: node.InstanceName},
		cli.DescribeRow{Attribute: prefix + ".InstanceType", Content: node.InstanceType},
		cli.DescribeRow{Attribute: prefix + ".KubeProxy", Content: node.KubeProxy.Mode},
		cli.DescribeRow{Attribute: prefix + ".MachineType", Content: node.MachineType},
		cli.DescribeRow{Attribute: prefix + ".Memory", Content: fmt.Sprintf("%d", node.Memory)},
		cli.DescribeRow{Attribute: prefix + ".NodeId", Content: node.NodeId},
		cli.DescribeRow{Attribute: prefix + ".NodeLogInfo", Content: node.NodeLogInfo},
		cli.DescribeRow{Attribute: prefix + ".NodeRole", Content: node.NodeRole},
		cli.DescribeRow{Attribute: prefix + ".NodeStatus", Content: node.NodeStatus},
		cli.DescribeRow{Attribute: prefix + ".OsName", Content: node.OsName},
		cli.DescribeRow{Attribute: prefix + ".OsType", Content: node.OsType},
		cli.DescribeRow{Attribute: prefix + ".Unschedulable", Content: fmt.Sprintf("%t", node.Unschedulable)},
		cli.DescribeRow{Attribute: prefix + ".Zone", Content: node.Zone},
	)
	for i, ip := range node.IPSet {
		ipPrefix := fmt.Sprintf("%s.IPSet[%d]", prefix, i)
		rows = append(rows,
			cli.DescribeRow{Attribute: ipPrefix + ".IP", Content: ip.IP},
			cli.DescribeRow{Attribute: ipPrefix + ".IPId", Content: ip.IPId},
			cli.DescribeRow{Attribute: ipPrefix + ".Default", Content: ip.Default},
			cli.DescribeRow{Attribute: ipPrefix + ".Mac", Content: ip.Mac},
			cli.DescribeRow{Attribute: ipPrefix + ".SubnetId", Content: ip.SubnetId},
			cli.DescribeRow{Attribute: ipPrefix + ".Type", Content: ip.Type},
			cli.DescribeRow{Attribute: ipPrefix + ".VPCId", Content: ip.VPCId},
			cli.DescribeRow{Attribute: ipPrefix + ".Bandwidth", Content: fmt.Sprintf("%d", ip.Bandwidth)},
		)
	}
	return rows
}

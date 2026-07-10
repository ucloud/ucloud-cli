package uk8s

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newNodeDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewDescribeUK8SNodeRequest()

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of a UK8S node",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			*req.Name = ctx.PickResourceID(*req.Name)
			node, err := client.DescribeUK8SNode(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []cli.DescribeRow{
				{Attribute: "Name", Content: node.Name},
				{Attribute: "Hostname", Content: node.Hostname},
				{Attribute: "InternalIP", Content: node.InternalIP},
				{Attribute: "ProviderID", Content: node.ProviderID},
				{Attribute: "CPUCapacity", Content: node.CPUCapacity},
				{Attribute: "MemoryCapacity", Content: node.MemoryCapacity},
				{Attribute: "PodCapacity", Content: fmt.Sprintf("%d", node.PodCapacity)},
				{Attribute: "AllocatedPods", Content: fmt.Sprintf("%d", node.AllocatedPodCount)},
				{Attribute: "Unschedulable", Content: fmt.Sprintf("%t", node.Unschedulable)},
				{Attribute: "KubeletVersion", Content: node.KubeletVersion},
				{Attribute: "KubeProxyVersion", Content: node.KubeProxyVersion},
				{Attribute: "ContainerRuntime", Content: node.ContainerRuntimeVersion},
				{Attribute: "OSImage", Content: node.OSImage},
				{Attribute: "KernelVersion", Content: node.KernelVersion},
				{Attribute: "Labels", Content: strings.Join(node.Labels, ",")},
				{Attribute: "Taints", Content: strings.Join(node.Taints, ",")},
				{Attribute: "Created", Content: time.Unix(int64(node.CreationTimestamp), 0).Format(time.RFC3339)},
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterId = flags.String("cluster-id", "", "Required. Cluster ID.")
	req.Name = flags.String("node-id", "", "Required. Node ID or IP.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("cluster-id")
	cmd.MarkFlagRequired("node-id")
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, nil, derefStr(req.Region), derefStr(req.ProjectId))
	})
	command.SetCompletion(cmd, "node-id", func() []string {
		return listNodeIDs(ctx, derefStr(req.ClusterId), derefStr(req.ProjectId), derefStr(req.Region))
	})
	return cmd
}

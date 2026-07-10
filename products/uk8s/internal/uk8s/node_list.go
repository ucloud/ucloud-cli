package uk8s

import (
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
			rows := make([]nodeRow, 0, len(resp.NodeSet))
			for _, node := range resp.NodeSet {
				rows = append(rows, nodeRow{
					ResourceID: node.NodeId, InstanceID: node.InstanceId, Name: node.InstanceName,
					Role: node.NodeRole, Zone: node.Zone, MachineType: node.MachineType,
					CPU: node.CPU, MemoryMB: node.Memory, Status: node.NodeStatus, OS: node.OsName,
				})
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

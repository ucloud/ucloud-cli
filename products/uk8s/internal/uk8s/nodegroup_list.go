package uk8s

import (
	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newNodeGroupList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewListUK8SNodeGroupRequest()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UK8S node groups",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			resp, err := client.ListUK8SNodeGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]nodeGroupRow, 0, len(resp.NodeGroupList))
			for _, group := range resp.NodeGroupList {
				rows = append(rows, nodeGroupRow{
					ResourceID: group.NodeGroupId, Name: group.NodeGroupName,
					MachineType: group.MachineType, CPU: group.CPU, MemoryMB: group.Mem,
					NodeCount: len(group.NodeList), ChargeType: group.ChargeType, ImageID: group.ImageId,
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

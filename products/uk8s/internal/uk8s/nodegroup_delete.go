package uk8s

import (
	"fmt"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newNodeGroupDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewRemoveUK8SNodeGroupRequest()
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a UK8S node group",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the UK8S node group?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			*req.NodeGroupId = ctx.PickResourceID(*req.NodeGroupId)
			if _, err := client.RemoveUK8SNodeGroup(req); err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "uk8s nodegroup[%s] deletion requested\n", *req.NodeGroupId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.NodeGroupId, Action: "delete", Status: "Deleting"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterId = flags.String("cluster-id", "", "Required. Cluster ID.")
	req.NodeGroupId = flags.String("nodegroup-id", "", "Required. Node group ID.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("cluster-id")
	cmd.MarkFlagRequired("nodegroup-id")
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, nil, derefStr(req.Region), derefStr(req.ProjectId))
	})
	command.SetCompletion(cmd, "nodegroup-id", func() []string {
		return listNodeGroupIDs(ctx, derefStr(req.ClusterId), derefStr(req.ProjectId), derefStr(req.Region))
	})
	return cmd
}

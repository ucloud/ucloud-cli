package uk8s

import (
	"fmt"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newNodeDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewDelUK8SClusterNodeV2Request()
	var nodeIDs []string
	var releaseDataUDisk bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete nodes from a UK8S cluster",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the UK8S node(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			results := make([]cli.OpResultRow, 0, len(nodeIDs))
			for _, idName := range nodeIDs {
				id := ctx.PickResourceID(idName)
				req.NodeId = sdk.String(id)
				req.ReleaseDataUDisk = sdk.Bool(releaseDataUDisk)
				if _, err := client.DelUK8SClusterNodeV2(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "uk8s node[%s] deletion requested\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleting"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterId = flags.String("cluster-id", "", "Required. Cluster ID.")
	flags.StringSliceVar(&nodeIDs, "node-id", nil, "Required. Node ID(s).")
	flags.BoolVar(&releaseDataUDisk, "release-data-udisk", true, "Optional. Release data disks.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")
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

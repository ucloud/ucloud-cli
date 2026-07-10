package uk8s

import (
	"fmt"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewDelUK8SClusterRequest()

	var clusterIDs []string
	var releaseUDisk bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete UK8S clusters",
		Long:  "Delete one or more UK8S clusters by cluster ID.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the UK8S cluster(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}

			w := ctx.ProgressWriter()
			results := make([]cli.OpResultRow, 0, len(clusterIDs))
			for _, idName := range clusterIDs {
				id := ctx.PickResourceID(idName)
				req.ClusterId = sdk.String(id)
				req.ReleaseUDisk = sdk.Bool(releaseUDisk)
				if _, err := client.DelUK8SCluster(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "uk8s[%s] deletion requested\n", id)
				results = append(results, cli.OpResultRow{
					ResourceID: id,
					Action:     "delete",
					Status:     "Deleting",
				})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&clusterIDs, "cluster-id", nil, "Required. Cluster ID(s) to delete.")
	flags.BoolVar(&releaseUDisk, "release-udisk", false, "Optional. Release data disks attached to cluster nodes.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("cluster-id")
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, nil, derefStr(req.Region), derefStr(req.ProjectId))
	})
	return cmd
}

package eip

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newModifyBandwidth ucloud eip modify-bw
func newModifyBandwidth(ctx *cli.Context) *cobra.Command {
	ids := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewModifyEIPBandwidthRequest()
	cmd := &cobra.Command{
		Use:     "modify-bw",
		Short:   "Modify bandwith of EIP instances",
		Long:    "Modify bandwith of EIP instances",
		Example: "ucloud eip modify-bw --eip-id eip-xx1,eip-xx2 --bandwidth-mb 20",
		// Deprecated: "use 'ucloud eip modiy'",
		Run: func(cmd *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, id := range ids {
				id = ctx.PickResourceID(id)
				req.EIPId = &id
				_, err := client.ModifyEIPBandwidth(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(ctx.ProgressWriter(), "eip[%s]'s bandwidth modified\n", id)
					results = append(results, cli.OpResultRow{ResourceID: id, Action: "modify-bw", Status: "Modified"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringSliceVarP(&ids, "eip-id", "", nil, "Required, Resource ID of EIPs to modify bandwidth")
	req.Bandwidth = cmd.Flags().Int("bandwidth-mb", 0, "Required. Bandwidth of EIP after modifed. Charge by traffic, range [1,300]; charge by bandwidth, range [1,800]")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, nil, nil)
	})
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("bandwidth-mb")
	return cmd
}

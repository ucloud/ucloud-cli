package bw

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newSharedDelete returns ucloud bw shared delete.
func newSharedDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewReleaseShareBandwidthRequest()
	ids := []string{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete shared bandwidth instance",
		Long:  "Delete shared bandwidth instance",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range ids {
				id := ctx.PickResourceID(idname)
				req.ShareBandwidthId = sdk.String(id)
				_, err := client.ReleaseShareBandwidth(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "shared bandwidth[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, "shared-bw-id", nil, "Required. Resource ID of shared bandwidth instances to delete")
	req.EIPBandwidth = flags.Int("eip-bandwidth-mb", 1, "Optional. Bandwidth of the joined EIPs,after deleting the shared bandwidth instance")
	req.PayMode = flags.String("traffic-mode", "", "Optional. The charge mode of joined EIPs after deleting the shared bandwidth. Accept values:Bandwidth,Traffic")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	command.SetCompletion(cmd, "shared-bw-id", func() []string {
		list, _ := getAllSharedBW(ctx, *req.ProjectId, *req.Region)
		return list
	})
	command.SetFlagValues(cmd, "traffic-mode", "Bandwidth", "Traffic")

	cmd.MarkFlagRequired("shared-bw-id")

	return cmd
}

package udpn

import (
	"fmt"

	"github.com/spf13/cobra"

	udpnsdk "github.com/ucloud/ucloud-sdk-go/services/udpn"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newModifyBW ucloud udpn modify-bw
func newModifyBW(ctx *cli.Context) *cobra.Command {
	idNames := []string{}
	client := cli.NewServiceClient(ctx, udpnsdk.NewClient)
	req := client.NewModifyUDPNBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "modify-bw",
		Short: "Modify bandwidth of UDPN tunnel",
		Long:  "Modify bandwidth of UDPN tunnel",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.UDPNId = sdk.String(id)
				_, err := client.ModifyUDPNBandwidth(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "udpn[%s]'s bandwidth modified\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "modify-bw", Status: "Modified"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udpn-id", nil, "Required. Resource ID of UDPN to modify bandwidth")
	req.Bandwidth = flags.Int("bandwidth-mb", 0, "Required. Bandwidth of UDPN tunnel. Unit:Mb. Range [2,1000]")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	ctx.SetCompletion(cmd, "udpn-id", func() []string {
		return getAllUDPNIdNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("udpn-id")
	cmd.MarkFlagRequired("bandwidth-mb")

	return cmd
}

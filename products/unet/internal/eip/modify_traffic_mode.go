package eip

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newSetChargeMode ucloud eip modify-traffic-mode
func newSetChargeMode(ctx *cli.Context) *cobra.Command {
	ids := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewSetEIPPayModeRequest()
	cmd := &cobra.Command{
		Use:     "modify-traffic-mode",
		Short:   "Modify charge mode of EIP instances",
		Long:    "Modify charge mode of EIP instances",
		Example: "ucloud eip modify-traffic-mode --eip-id eip-xx1,eip-xx2 --traffic-mode Traffic",
		Run: func(cmd *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, id := range ids {
				id = ctx.PickResourceID(id)
				req.EIPId = &id
				eipIns, err := getEIP(ctx, id)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.Bandwidth = sdk.Int(eipIns.Bandwidth)
				_, err = client.SetEIPPayMode(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(ctx.ProgressWriter(), "eip[%s]'s charge mode was modified to %s\n", id, *req.PayMode)
					results = append(results, cli.OpResultRow{ResourceID: id, Action: "modify-traffic-mode", Status: "Modified"})
				}
			}
			ctx.EmitResult(results...)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringSliceVarP(&ids, "eip-id", "", nil, "Required, Resource ID of EIPs to modify charge mode")
	req.PayMode = cmd.Flags().String("traffic-mode", "", "Required, Charge mode of eip, 'Traffic','Bandwidth' or 'PostAccurateBandwidth'")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	command.SetFlagValues(cmd, "traffic-mode", "Bandwidth", "Traffic", "PostAccurateBandwidth")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, nil, nil)
	})
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("traffic-mode")
	return cmd
}

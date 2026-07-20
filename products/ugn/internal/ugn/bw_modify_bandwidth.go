package ugn

import (
	"fmt"

	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newBWModifyBandwidth ucloud ugn bw modify-bandwidth
func newBWModifyBandwidth(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewModifyUGNBandwidthRequest()
	cmd := &cobra.Command{
		Use:   "modify-bandwidth",
		Short: "Modify ugn bandwidth",
		Long:  "Modify ugn bandwidth",
		Run: func(c *cobra.Command, args []string) {
			_, err := client.ModifyUGNBandwidth(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "ugn bw[%s] bandwidth modified to %d\n", *req.PackageID, *req.BandWidth)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.PackageID, Action: "modify-bandwidth", Status: "Modified"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.BandWidth = flags.Int("bandwidth", 0, "Required. New bandwidth value")
	req.PackageID = flags.String("package-id", "", "Required. Bandwidth package ID")
	req.UGNID = flags.String("ugn-id", "", "Required. Resource ID of the ugn instance")

	cmd.MarkFlagRequired("bandwidth")
	cmd.MarkFlagRequired("package-id")
	cmd.MarkFlagRequired("ugn-id")

	return cmd
}

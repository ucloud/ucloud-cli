package bw

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPkgList returns ucloud bw pkg list.
func newPkgList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeBandwidthPackageRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bandwidth packages",
		Long:  "List bandwidth packages",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeBandwidthPackage(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []BandwidthPkgRow{}
			for _, bp := range resp.DataSets {
				row := BandwidthPkgRow{
					ResourceID: bp.BandwidthPackageId,
					Bandwidth:  strconv.Itoa(bp.Bandwidth) + "MB",
					StartTime:  common.FormatDateTime(bp.EnableTime),
					EndTime:    common.FormatDateTime(bp.DisableTime),
				}
				eip := bp.EIPId
				for _, addr := range bp.EIPAddr {
					eip += "/" + addr.IP + "/" + addr.OperatorName
				}
				row.EIP = eip
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit range [0,10000000]")

	return cmd
}

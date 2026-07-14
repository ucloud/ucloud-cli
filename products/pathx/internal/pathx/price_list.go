package pathx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPriceList ucloud pathx price list
func newPriceList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	priceReq := client.NewGetUGA3PriceRequest()
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all the pathx acceleration area price",
		Long:    "List all the pathx acceleration area price",
		Example: "ucloud pathx price list --bandwidth 10 --area-code BKK --charge-type Month",
		Run: func(cmd *cobra.Command, args []string) {
			if strings.EqualFold(*priceReq.ChargeType, "Month") {
				*priceReq.Quantity = 0
			} else if *priceReq.Quantity <= 0 {
				ctx.HandleError(fmt.Errorf("If the value of charge-type is 'Year' or 'Hour',its value must be greater than 0"))
				return
			}
			switch strings.ToLower(*priceReq.ChargeType) {
			case "hour":
				*priceReq.ChargeType = "Dynamic"
			case "month":
				*priceReq.ChargeType = "Month"
			case "year":
				*priceReq.ChargeType = "Year"
			}
			response, err := client.GetUGA3Price(priceReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			priceList := response.UGA3Price
			if len(priceList) == 0 {
				ctx.HandleError(fmt.Errorf("Not found acceleration area price information."))
				return
			}
			list := make([]UGA3PriceRow, 0)
			for _, info := range priceList {
				list = append(list, UGA3PriceRow{
					AccelerationBandwidthPrice: fmt.Sprintf("%s%s", "￥", strconv.FormatFloat(info.AccelerationBandwidthPrice, 'g', 12, 64)),
					AccelerationForwarderPrice: fmt.Sprintf("%s%s", "￥", strconv.FormatFloat(info.AccelerationForwarderPrice, 'g', 12, 64)),
					AccelerationArea:           info.AccelerationArea,
				})
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindProjectID(cmd, priceReq)
	ctx.BindRegion(cmd, priceReq)
	ctx.BindZone(cmd, priceReq)
	priceReq.Bandwidth = flags.Int("bandwidth", 1, "Required. The bandwidth of acceleration area to get price")
	priceReq.AreaCode = flags.String("area-code", "", "Required. The area-code of acceleration area to get price")
	priceReq.Quantity = flags.Int("quantity", 1, "Optional. When the value of the charge-type is 'Month',its default value is 0,if the value of charge-type is 'Year' or 'Hour',its value must be greater than 0")
	priceReq.ChargeType = flags.String("charge-type", "", "Optional. Its value is not case sensitive,acceptable values:'Year',pay yearly;'Month',pay monthly;'Hour',pay hourly")
	priceReq.AccelerationArea = flags.String("accel", "", "Optional. The acceleration-area to get price")
	_ = cmd.MarkFlagRequired("bandwidth")
	_ = cmd.MarkFlagRequired("area-code")
	command.SetFlagValues(cmd, "area-code", "BKK", "DXB", "FRA", "SGN", "HKG", "CGK", "LOS", "LHR", "LAX", "MNL", "DME", "BOM", "MSP", "ICN", "PVG", "SIN", "NRT", "IAD", "TPE")
	command.SetFlagValues(cmd, "charge-type", "Year", "Month", "Hour")
	command.SetFlagValues(cmd, "accel", "Global", "AP", "EU", "ME", "OA", "AF", "NA", "SA")
	return cmd
}

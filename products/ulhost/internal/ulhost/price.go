package ulhost

import (
	"fmt"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPrice ucloud ulhost price
func newPrice(ctx *cli.Context) *cobra.Command {
	var renew bool
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	cmd := &cobra.Command{
		Use:   "price",
		Short: "Get ULHost instance price",
		Long:  `Get ULHost instance price for creating or renewing`,
		Run: func(cmd *cobra.Command, args []string) {
			if renew {
				showRenewPrice(ctx, client, cmd)
			} else {
				showCreatePrice(ctx, client, cmd)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.String("bundle-id", "", "Required for create price. Bundle ID of the ULHost instance")
	flags.String("charge-type", "", "Optional. 'Year' or 'Month'. If not specified, return all charge types")
	flags.Int("count", 1, "Optional. Number of instances. Range [1,5]")
	flags.Int("quantity", 1, "Optional. Purchase duration. Default: 1")
	flags.String("ulhost-id", "", "Required for renew price. ULHost instance ID")
	flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	flags.BoolVar(&renew, "renew", false, "Optional. Get renew price instead of create price")
	command.SetFlagValues(cmd, "charge-type", "Month", "Year")
	command.SetCompletion(cmd, "region", ctx.RegionList)

	return cmd
}

func showCreatePrice(ctx *cli.Context, client *ucompsharesdk.UCompShareClient, cmd *cobra.Command) {
	req := client.NewGetULHostInstancePriceRequest()
	bundleID, _ := cmd.Flags().GetString("bundle-id")
	chargeType, _ := cmd.Flags().GetString("charge-type")
	count, _ := cmd.Flags().GetInt("count")
	quantity, _ := cmd.Flags().GetInt("quantity")
	projectID, _ := cmd.Flags().GetString("project-id")
	region, _ := cmd.Flags().GetString("region")

	req.BundleId = sdk.String(bundleID)
	req.ChargeType = sdk.String(chargeType)
	req.Count = sdk.Int(count)
	req.Quantity = sdk.Int(quantity)
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)

	if bundleID == "" {
		ctx.HandleError(fmt.Errorf("--bundle-id is required for create price"))
		return
	}

	resp, err := client.GetULHostInstancePrice(req)
	if err != nil {
		ctx.HandleError(err)
		return
	}
	rows := make([]priceRow, 0, len(resp.PriceSet))
	for _, price := range resp.PriceSet {
		rows = append(rows, priceRow{
			ChargeType:    price.ChargeType,
			Price:         fmt.Sprintf("%.2f", price.Price),
			OriginalPrice: fmt.Sprintf("%.2f", price.OriginalPrice),
		})
	}
	ctx.PrintList(rows)
}

func showRenewPrice(ctx *cli.Context, client *ucompsharesdk.UCompShareClient, cmd *cobra.Command) {
	req := client.NewGetULHostRenewPriceRequest()
	ulhostID, _ := cmd.Flags().GetString("ulhost-id")
	chargeType, _ := cmd.Flags().GetString("charge-type")
	projectID, _ := cmd.Flags().GetString("project-id")
	region, _ := cmd.Flags().GetString("region")

	req.ULHostId = sdk.String(ulhostID)
	req.ChargeType = sdk.String(chargeType)
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)

	if ulhostID == "" {
		ctx.HandleError(fmt.Errorf("--ulhost-id is required for renew price"))
		return
	}

	resp, err := client.GetULHostRenewPrice(req)
	if err != nil {
		ctx.HandleError(err)
		return
	}
	rows := make([]priceRow, 0, len(resp.PriceSet))
	for _, price := range resp.PriceSet {
		rows = append(rows, priceRow{
			ChargeType:    price.ChargeType,
			Price:         fmt.Sprintf("%.2f", price.Price),
			OriginalPrice: fmt.Sprintf("%.2f", price.OriginalPrice),
		})
	}
	ctx.PrintList(rows)
}

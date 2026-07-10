package service

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPrice ucloud urocketmq service price
func newPrice(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewGetURocketMQServicePriceRequest()
	cmd := &cobra.Command{
		Use:   "price",
		Short: "Get price of URocketMQ service instance",
		Long:  "Get price of URocketMQ service instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if *req.Storage <= 0 || *req.Storage%100 != 0 {
				return fmt.Errorf("--storage-gb must be a positive multiple of 100")
			}
			resp, err := client.GetURocketMQServicePrice(req)
			if err != nil {
				return err
			}
			list := make([]urocketmq.PriceSet, 0, len(resp.PriceSet))
			for _, p := range resp.PriceSet {
				list = append(list, p)
			}
			if ctx.Format() != cli.OutputTable {
				ctx.PrintList(list)
				return nil
			}
			rows := make([]priceRowDefault, 0, len(list))
			for _, r := range list {
				rows = append(rows, priceRowDefault{
					ChargeName: r.ChargeName,
					ChargeType: r.ChargeType,
					Price:      r.Price,
				})
			}
			ctx.PrintList(rows)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ChargeType = flags.String("charge-type", "Month", "Required. Charge type. Enum: Year, Month, Dynamic")
	req.Edition = flags.String("edition", "Enterprise", "Required. Edition. Unique value: Enterprise")
	req.Mode = flags.String("mode", "PrivateNet", "Required. Network mode. Unique value: PrivateNet")
	req.PublicVersion = flags.String("public-version", "", "Required. Cluster version. Enum: v4, v5. Note: each region currently supports only one version, confirm with region")
	req.Quantity = flags.Int("quantity", 1, "Optional. Purchase duration. Dynamic does not require this; Month 0 means until month end")
	req.Storage = flags.Int("storage-gb", 0, "Required. Message storage space in GB")
	req.TPS = flags.Int("tps", 0, "Required. Transactions per second. Enum: 10000, 20000, 50000, 100000, 200000. Note: v4 supports 20000, 50000, 100000, 200000; v5 currently supports only 10000, 20000.")

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)

	command.SetFlagValues(cmd, "charge-type", "Year", "Month", "Dynamic")
	command.SetFlagValues(cmd, "edition", "Enterprise")
	command.SetFlagValues(cmd, "mode", "PrivateNet")
	command.SetFlagValues(cmd, "public-version", "v4", "v5")
	command.SetFlagValues(cmd, "tps", "10000", "20000", "50000", "100000", "200000")

	cmd.MarkFlagRequired("charge-type")
	cmd.MarkFlagRequired("edition")
	cmd.MarkFlagRequired("mode")
	cmd.MarkFlagRequired("public-version")
	cmd.MarkFlagRequired("storage-gb")
	cmd.MarkFlagRequired("tps")

	return cmd
}

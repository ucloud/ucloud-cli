package pgsql

import (
	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPrice ucloud pgsql db price
func newPrice(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewGetUPgSQLInstancePriceRequest()
	cmd := &cobra.Command{
		Use:   "price",
		Short: "Get the price of creating UPgSQL instances",
		Long:  "Get the price of creating UPgSQL instances",
		Run: func(c *cobra.Command, args []string) {
			if *req.ChargeType == "Dynamic" {
				req.Quantity = sdk.Int(0)
			}
			resp, err := client.GetUPgSQLInstancePrice(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []PgsqlPriceRow{}
			for _, p := range resp.PriceSet {
				rows = append(rows, PgsqlPriceRow{
					ChargeType:    p.ChargeType,
					Price:         p.Price,
					OriginalPrice: p.OriginalPrice,
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.MachineType = flags.String("machine-type", "", "Required. Machine type ID, e.g. o.pgsql2m.medium. See 'ucloud pgsql db list-machine-type'")
	req.DiskSpace = flags.Int("disk-size-gb", 0, "Required. Disk space (GiB)")
	req.InstanceMode = flags.String("mode", "Normal", "Required. Normal / HA")
	req.ChargeType = flags.String("charge-type", "Month", "Optional. Year / Month / Dynamic")
	req.Quantity = flags.Int("quantity", 1, "Optional. Purchase duration. Month: 1-9, 0=until end of month; Dynamic: ignored; Year: years")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	command.SetFlagValues(cmd, "charge-type", "Year", "Month", "Dynamic")
	command.SetFlagValues(cmd, "mode", "Normal", "HA")
	command.SetCompletion(cmd, "machine-type", func() []string {
		return listMachineTypeIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	cmd.MarkFlagRequired("machine-type")
	cmd.MarkFlagRequired("disk-size-gb")

	return cmd
}

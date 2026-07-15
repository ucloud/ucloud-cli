package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUpgradePrice ucloud pgsql db upgrade-price
func newUpgradePrice(ctx *cli.Context) *cobra.Command {
	client := newUPgSQLClient(ctx)
	req := client.NewGetUPgSQLUpgradePriceRequest()
	cmd := &cobra.Command{
		Use:   "upgrade-price",
		Short: "Get the price of upgrading a UPgSQL instance",
		Long:  "Get the price of upgrading disk space and/or machine type of a UPgSQL instance",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			resp, err := client.GetUPgSQLUpgradePrice(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.PrintList([]PgsqlPriceRow{{
				Price:         resp.Price,
				OriginalPrice: resp.OriginalPrice,
			}})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	req.MachineType = flags.String("machine-type", "", "Required. New machine type ID. See 'ucloud pgsql db list-machine-type'")
	req.DiskSpace = flags.Int("disk-size-gb", 0, "Required. New disk space (GiB)")
	req.InstanceMode = flags.String("mode", "Normal", "Optional. Normal / HA")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	command.SetFlagValues(cmd, "mode", "Normal", "HA")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})
	command.SetCompletion(cmd, "machine-type", func() []string {
		return listMachineTypeIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("machine-type")
	cmd.MarkFlagRequired("disk-size-gb")

	return cmd
}

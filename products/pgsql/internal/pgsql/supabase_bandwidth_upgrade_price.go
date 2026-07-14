package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseBandwidthUpgradePrice ucloud pgsql supabase bandwidth-upgrade-price
func newSupabaseBandwidthUpgradePrice(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var bandwidth int
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "bandwidth-upgrade-price",
		Short: "Get the price of upgrading external-access bandwidth of a USupabase instance",
		Long:  "Get the price of upgrading external-access bandwidth of a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			params["InstanceID"] = instanceID
			params["Bandwidth"] = bandwidth
			payload, err := invokeSupabase(ctx, "DescribeUSupabaseBandwithUpgradePrice", params)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []SupabaseChargeRow{}
			if ds, ok := payload["DataSet"].([]interface{}); ok {
				for _, item := range ds {
					m, _ := item.(map[string]interface{})
					rows = append(rows, SupabaseChargeRow{
						ChargeType: getString(m, "ChargeType"),
						Price:      getInt(m, "Price"),
					})
				}
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	flags.IntVar(&bandwidth, "bandwidth", 0, "Required. Target bandwidth (Mbps)")
	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("bandwidth")

	return cmd
}

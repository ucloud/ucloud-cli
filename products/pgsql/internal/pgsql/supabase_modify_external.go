package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseModifyExternal ucloud pgsql supabase modify-external-access
func newSupabaseModifyExternal(ctx *cli.Context) *cobra.Command {
	var instanceID, whiteList string
	var bandwidth, port int
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "modify-external-access",
		Short: "Modify external-access settings of a USupabase instance",
		Long:  "Modify external-access bandwidth/port/whitelist of a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			params["InstanceID"] = instanceID
			if c.Flags().Changed("bandwidth") {
				params["Bandwidth"] = bandwidth
			}
			if c.Flags().Changed("port") {
				params["Port"] = port
			}
			if c.Flags().Changed("white-list") {
				params["WhiteList"] = whiteList
			}
			if _, err := invokeSupabase(ctx, "ModifyUSupabaseExternalAccessInfo", params); err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "supabase[%s] external access modified\n", instanceID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "modify-external-access", Status: "Modified"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	flags.IntVar(&bandwidth, "bandwidth", 0, "Optional. New bandwidth (Mbps)")
	flags.IntVar(&port, "port", 0, "Optional. New external port")
	flags.StringVar(&whiteList, "white-list", "", "Optional. New IP whitelist")
	cmd.MarkFlagRequired("instance-id")

	return cmd
}

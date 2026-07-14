package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseEnableExternal ucloud pgsql supabase enable-external-access
func newSupabaseEnableExternal(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var bandwidth int
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "enable-external-access",
		Short: "Enable external access for a USupabase instance",
		Long:  "Enable external network access for a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			params["InstanceID"] = instanceID
			params["Bandwidth"] = bandwidth
			if _, err := invokeSupabase(ctx, "EnableUSupabaseExternalAccess", params); err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "supabase[%s] external access enabled\n", instanceID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "enable-external-access", Status: "Enabled"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	flags.IntVar(&bandwidth, "bandwidth", 0, "Required. Bandwidth (Mbps)")
	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("bandwidth")

	return cmd
}

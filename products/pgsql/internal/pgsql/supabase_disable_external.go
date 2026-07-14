package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseDisableExternal ucloud pgsql supabase disable-external-access
func newSupabaseDisableExternal(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "disable-external-access",
		Short: "Disable external access for a USupabase instance",
		Long:  "Disable external network access for a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			params["InstanceID"] = instanceID
			if _, err := invokeSupabase(ctx, "DisableUSupabaseExternalAccess", params); err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "supabase[%s] external access disabled\n", instanceID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "disable-external-access", Status: "Disabled"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	cmd.MarkFlagRequired("instance-id")

	return cmd
}

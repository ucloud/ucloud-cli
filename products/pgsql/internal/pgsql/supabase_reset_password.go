package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseResetPassword ucloud pgsql supabase reset-password
func newSupabaseResetPassword(ctx *cli.Context) *cobra.Command {
	var instanceID, password string
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset the dashboard password of a USupabase instance",
		Long:  "Reset the dashboard password of a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			params["InstanceID"] = instanceID
			params["Password"] = password
			if _, err := invokeSupabase(ctx, "ResetUSupabasePassword", params); err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "supabase[%s] password reset\n", instanceID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "reset-password", Status: "PasswordReset"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	flags.StringVar(&password, "password", "", "Required. New dashboard password")
	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("password")

	return cmd
}

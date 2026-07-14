package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseRestart ucloud pgsql supabase restart
func newSupabaseRestart(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var async bool
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart a USupabase instance",
		Long:  "Restart a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			params["InstanceID"] = instanceID
			if _, err := invokeSupabase(ctx, "RestartUSupabase", params); err != nil {
				ctx.HandleError(err)
				return
			}
			w := ctx.ProgressWriter()
			if async {
				fmt.Fprintf(w, "supabase[%s] is restarting\n", instanceID)
			} else {
				text := fmt.Sprintf("supabase[%s] is restarting", instanceID)
				ctx.PollerTo(w, describeSupabaseByID(ctx, common.region, common.zone, common.projectID, common.memoryDB)).
					Spoll(instanceID, text, []string{SUPABASE_RUNNING, SUPABASE_FAIL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "restart", Status: "Restarting"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish")
	cmd.MarkFlagRequired("instance-id")

	return cmd
}

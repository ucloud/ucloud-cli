package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseDelete ucloud pgsql supabase delete
func newSupabaseDelete(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var yes bool
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a USupabase instance",
		Long:  "Delete a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure you want to delete supabase[%s]?", instanceID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			params := common.params()
			params["InstanceID"] = instanceID
			if _, err := invokeSupabase(ctx, "DeleteUSupabase", params); err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "supabase[%s] deleted\n", instanceID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "delete", Status: "Deleted"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation")
	cmd.MarkFlagRequired("instance-id")

	return cmd
}

package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseGetAPIKey ucloud pgsql supabase get-api-key
func newSupabaseGetAPIKey(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "get-api-key",
		Short: "Display the API keys (service key + anon key) of a USupabase instance",
		Long:  "Display the API keys (service key + anon key) of a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			params["InstanceID"] = instanceID
			payload, err := invokeSupabase(ctx, "GetUSupabaseAPIKey", params)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			key, _ := payload["Key"].(map[string]interface{})
			ctx.PrintList([]SupabaseAPIKeyRow{{
				ServiceKey: getString(key, "ServiceKey"),
				AnonKey:    getString(key, "AnonKey"),
			}})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	cmd.MarkFlagRequired("instance-id")

	return cmd
}

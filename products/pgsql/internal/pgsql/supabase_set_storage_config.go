package pgsql

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseSetStorageConfig ucloud pgsql supabase set-storage-config
func newSupabaseSetStorageConfig(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var configs []string
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "set-storage-config",
		Short: "Set storage configuration entries of a USupabase instance",
		Long:  "Set storage configuration entries of a USupabase instance (use --config key=value, repeatable)",
		Run: func(c *cobra.Command, args []string) {
			entries := []map[string]interface{}{}
			for _, kv := range configs {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) != 2 {
					ctx.HandleError(fmt.Errorf("invalid --config %q, expected key=value", kv))
					return
				}
				entries = append(entries, map[string]interface{}{
					"Key":   parts[0],
					"Value": parts[1],
				})
			}
			params := common.params()
			params["InstanceID"] = instanceID
			params["ConfigEntry"] = entries
			if _, err := invokeSupabase(ctx, "SetUSupabaseStorageConfig", params); err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "supabase[%s] storage config set\n", instanceID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "set-storage-config", Status: "Set"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	flags.StringSliceVar(&configs, "config", nil, "Required. Storage config entry, format key=value (repeatable)")
	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("config")

	return cmd
}

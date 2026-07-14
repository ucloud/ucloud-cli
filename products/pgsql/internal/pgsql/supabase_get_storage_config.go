package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseGetStorageConfig ucloud pgsql supabase get-storage-config
func newSupabaseGetStorageConfig(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "get-storage-config",
		Short: "Display the storage configuration of a USupabase instance",
		Long:  "Display the storage configuration of a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			params["InstanceID"] = instanceID
			payload, err := invokeSupabase(ctx, "GetUSupabaseStorageConfig", params)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []SupabaseStorageConfigRow{}
			if ds, ok := payload["DataSet"].([]interface{}); ok {
				for _, item := range ds {
					m, _ := item.(map[string]interface{})
					rows = append(rows, SupabaseStorageConfigRow{
						Key:         getString(m, "Key"),
						Value:       getString(m, "Value"),
						Description: getString(m, "Description"),
						Required:    getBool(m, "Required"),
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
	cmd.MarkFlagRequired("instance-id")

	return cmd
}

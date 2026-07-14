package pgsql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseDescribe ucloud pgsql supabase describe
func newSupabaseDescribe(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Display details of a USupabase instance",
		Long:  "Display details of a USupabase instance",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			params["InstanceID"] = instanceID
			payload, err := invokeSupabase(ctx, "DescribeUSupabase", params)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ds, ok := payload["DataSet"].(map[string]interface{})
			if !ok {
				ctx.HandleError(fmt.Errorf("pgsql supabase[%s] may not exist", instanceID))
				return
			}
			attrs := []cli.DescribeRow{
				{Attribute: "InstanceID", Content: getString(ds, "InstanceID")},
				{Attribute: "USupabaseName", Content: getString(ds, "USupabaseName")},
				{Attribute: "State", Content: getString(ds, "State")},
				{Attribute: "Zone", Content: getString(ds, "Zone")},
				{Attribute: "UPgSQLID", Content: getString(ds, "UPgSQLID")},
				{Attribute: "VPCID", Content: getString(ds, "VPCID")},
				{Attribute: "SubnetID", Content: getString(ds, "SubnetID")},
				{Attribute: "IntranetAddress", Content: getString(ds, "IntranetAddress")},
				{Attribute: "Port", Content: strconv.Itoa(getInt(ds, "Port"))},
				{Attribute: "ExternalNetworkStatus", Content: getString(ds, "ExternalNetworkStatus")},
				{Attribute: "ExternalNetworkAddress", Content: getString(ds, "ExternalNetworkAddress")},
				{Attribute: "ExternalNetworkPort", Content: strconv.Itoa(getInt(ds, "ExternalNetworkPort"))},
				{Attribute: "Bandwidth", Content: strconv.Itoa(getInt(ds, "Bandwidth"))},
				{Attribute: "WhiteList", Content: getString(ds, "WhiteList")},
			}
			fmt.Fprintln(ctx.ProgressWriter(), "Attributes:")
			ctx.PrintList(attrs)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Required. Resource ID of the USupabase instance")
	cmd.MarkFlagRequired("instance-id")

	return cmd
}

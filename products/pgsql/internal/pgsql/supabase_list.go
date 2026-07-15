package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// supabaseCommon holds the region/zone/project-id/memory-db flags shared by
// every supabase command. Bound via bindSupabaseCommon so flag order stays
// consistent across verbs (the completion golden depends on it).
type supabaseCommon struct {
	region    string
	zone      string
	projectID string
	memoryDB  bool
}

// bindSupabaseCommon registers --region/--zone/--project-id/--memory-db with
// ctx defaults + completion. ProjectId is ALWAYS bound: USupabase IAM checks
// require the ProjectId context, so every action carries it.
func bindSupabaseCommon(cmd *cobra.Command, ctx *cli.Context) *supabaseCommon {
	c := &supabaseCommon{}
	flags := cmd.Flags()
	flags.StringVar(&c.region, "region", ctx.DefaultRegion(), "Optional. Override default region, see 'ucloud region'")
	flags.StringVar(&c.zone, "zone", ctx.DefaultZone(), "Optional. Override default zone, see 'ucloud region'")
	flags.StringVar(&c.projectID, "project-id", ctx.DefaultProjectID(), "Optional. Override default project-id, see 'ucloud project list'")
	flags.BoolVar(&c.memoryDB, "memory-db", false, "Optional. Operate on the MemoryDB (AI memory) variant")
	command.SetCompletion(cmd, "region", func() []string { return ctx.RegionList() })
	command.SetCompletion(cmd, "zone", func() []string { return ctx.ZoneList(c.region) })
	command.SetCompletion(cmd, "project-id", func() []string { return ctx.ProjectList() })
	return c
}

// params builds the map payload common to every supabase action: Region/Zone/
// ProjectId/IsMemoryDB. ProjectId is mandatory (IAM context). Business fields
// are added by the caller.
func (c *supabaseCommon) params() map[string]interface{} {
	p := map[string]interface{}{
		"Region":    c.region,
		"Zone":      c.zone,
		"ProjectId": c.projectID,
	}
	if c.memoryDB {
		p["IsMemoryDB"] = true
	}
	return p
}

// newSupabaseList ucloud pgsql supabase list
func newSupabaseList(ctx *cli.Context) *cobra.Command {
	var instanceID string
	var limit, offset int
	var common *supabaseCommon
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List USupabase instances",
		Long:  "List USupabase instances (or MemoryDB instances with --memory-db)",
		Run: func(c *cobra.Command, args []string) {
			params := common.params()
			if instanceID != "" {
				params["InstanceID"] = instanceID
			}
			if limit > 0 {
				params["Limit"] = limit
			}
			if offset > 0 {
				params["Offset"] = offset
			}
			payload, err := invokeSupabase(ctx, "ListUSupabaseInstance", params)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []SupabaseInstanceRow{}
			if ds, ok := payload["DataSet"].([]interface{}); ok {
				for _, item := range ds {
					m, ok := item.(map[string]interface{})
					if !ok {
						continue
					}
					rows = append(rows, SupabaseInstanceRow{
						USupabaseName:   getString(m, "USupabaseName"),
						InstanceID:      getString(m, "InstanceID"),
						UPgSQLID:        getString(m, "UPgSQLID"),
						Zone:            getString(m, "Zone"),
						IntranetAddress: getString(m, "IntranetAddress"),
						Port:            getInt(m, "Port"),
						State:           getString(m, "State"),
					})
				}
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	common = bindSupabaseCommon(cmd, ctx)
	flags.StringVar(&instanceID, "instance-id", "", "Optional. List only the specified USupabase instance")
	flags.IntVar(&limit, "limit", 0, "Optional. Max instances per page (0 = default)")
	flags.IntVar(&offset, "offset", 0, "Optional. Offset")

	return cmd
}

// getString / getInt are tiny helpers over a generic response map (JSON numbers
// unmarshal as float64).
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSupabaseCreateMemoryDB ucloud pgsql supabase create-memory-db
//
// The MemoryDB (AI memory) variant uses the same business fields as create but
// the dedicated CreateUMemoryDB action (CreateUSupabaseRequest has no IsMemoryDB
// field, so the two are distinguished by action name, not a flag).
func newSupabaseCreateMemoryDB(ctx *cli.Context) *cobra.Command {
	var async bool
	var f *supabaseCreateFlags
	cmd := &cobra.Command{
		Use:   "create-memory-db",
		Short: "Create a UMemoryDB (AI memory) instance",
		Long:  "Create a UMemoryDB (AI memory) instance — the Supabase-based AI memory variant",
		Run: func(c *cobra.Command, args []string) {
			runSupabaseCreate(ctx, "CreateUMemoryDB", f, async)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	f = bindSupabaseCreate(cmd, ctx)
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish")

	for _, req := range []string{"name", "dashboard-name", "dashboard-password", "pgsql-user", "db-version", "param-group-id", "pgsql-password", "disk-size-gb", "machine-type", "subnet-id", "vpc-id"} {
		cmd.MarkFlagRequired(req)
	}
	return cmd
}

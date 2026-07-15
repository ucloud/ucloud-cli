package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPgsqlSupabase ucloud pgsql supabase
//
// USupabase is a pgsql-attached enhancement product (a community Supabase stack
// deployed onto a UPgSQL host). Its gateway actions are not in ucloud-sdk-go, so
// they are invoked generically (see supabase_client.go). The MemoryDB (AI memory)
// variant is the same backend toggled by --memory-db; create has its own action.
func newPgsqlSupabase(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supabase",
		Short: "Manage USupabase instances (and the MemoryDB AI-memory variant)",
		Long:  "Manage USupabase instances (and the MemoryDB AI-memory variant)",
	}
	cmd.AddCommand(newSupabaseList(ctx))
	cmd.AddCommand(newSupabaseDescribe(ctx))
	cmd.AddCommand(newSupabaseCreate(ctx))
	cmd.AddCommand(newSupabaseCreateMemoryDB(ctx))
	cmd.AddCommand(newSupabaseDelete(ctx))
	cmd.AddCommand(newSupabaseStart(ctx))
	cmd.AddCommand(newSupabaseStop(ctx))
	cmd.AddCommand(newSupabaseRestart(ctx))
	cmd.AddCommand(newSupabaseResetPassword(ctx))
	cmd.AddCommand(newSupabaseGetAPIKey(ctx))
	cmd.AddCommand(newSupabaseGetStorageConfig(ctx))
	cmd.AddCommand(newSupabaseSetStorageConfig(ctx))
	cmd.AddCommand(newSupabaseEnableExternal(ctx))
	cmd.AddCommand(newSupabaseDisableExternal(ctx))
	cmd.AddCommand(newSupabaseModifyExternal(ctx))
	cmd.AddCommand(newSupabaseExternalPrice(ctx))
	cmd.AddCommand(newSupabaseBandwidthUpgradePrice(ctx))
	return cmd
}

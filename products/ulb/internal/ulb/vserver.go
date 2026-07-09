package ulb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newVServer returns ucloud ulb vserver.
func newVServer(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vserver",
		Short: "List and manipulate ULB Vserver instances",
		Long:  "List and manipulate ULB Vserver instances",
	}
	cmd.AddCommand(newVServerList(ctx))
	cmd.AddCommand(newVServerCreate(ctx))
	cmd.AddCommand(newVServerUpdate(ctx))
	cmd.AddCommand(newVServerDelete(ctx))
	cmd.AddCommand(newBackend(ctx))
	cmd.AddCommand(newPolicy(ctx))
	return cmd
}

package nlb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand assembles the `nlb` command tree. This aggregator only constructs
// the top-level command and AddCommand's one constructor per verb / sub-group
// (§2.2 file-layout convention): no business logic lives here.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nlb",
		Short: "List and manipulate NLB (Network Load Balancer) instances",
		Long:  "List and manipulate NLB (Network Load Balancer) instances",
	}

	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newUpdate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newListener(ctx))
	cmd.AddCommand(newTarget(ctx))

	return cmd
}

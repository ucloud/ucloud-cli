package nlb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newTarget assembles the `nlb target` sub-tree (backend service nodes attached
// to a listener).
func newTarget(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target",
		Short: "Manage NLB backend targets (service nodes)",
		Long:  "Add, remove and update the backend targets of an NLB listener.",
	}
	cmd.AddCommand(newTargetAdd(ctx))
	cmd.AddCommand(newTargetRemove(ctx))
	cmd.AddCommand(newTargetUpdate(ctx))
	return cmd
}

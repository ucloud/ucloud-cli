package ulb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPolicy returns ucloud ulb vserver policy.
func newPolicy(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "List and manipulate forward policy for VServer",
		Long:  "List and manipulate forward policy for VServer",
	}
	cmd.AddCommand(newPolicyAdd(ctx))
	cmd.AddCommand(newPolicyList(ctx))
	cmd.AddCommand(newPolicyUpdate(ctx))
	cmd.AddCommand(newPolicyDelete(ctx))
	return cmd
}

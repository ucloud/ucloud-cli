package ulb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newBackend returns ucloud ulb vserver backend.
func newBackend(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backend",
		Short: "List and manipulate VServer backend nodes",
		Long:  "List and manipulate VServer backend nodes",
	}
	cmd.AddCommand(newBackendList(ctx))
	cmd.AddCommand(newBackendAdd(ctx))
	cmd.AddCommand(newBackendUpdate(ctx))
	cmd.AddCommand(newBackendDelete(ctx))
	return cmd
}

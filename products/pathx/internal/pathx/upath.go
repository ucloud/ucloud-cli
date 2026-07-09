package pathx

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUpath builds `ucloud pathx upath`.
func newUpath(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upath",
		Short: "List pathx upath instances",
		Long:  "List pathx upath instances",
	}
	cmd.AddCommand(newUpathList(ctx))
	return cmd
}

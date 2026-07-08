package pathx

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newArea builds `ucloud pathx area`.
func newArea(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "area",
		Short: "List origin area or acceleration area information",
		Long:  "List origin area or acceleration area information",
	}
	cmd.AddCommand(newAreaList(ctx))
	return cmd
}

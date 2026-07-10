package ext

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `ext` root command owned by products/eip.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ext",
		Short: "extended commands of UCloud CLI",
		Long:  "extended commands of UCloud CLI",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newUHost(ctx))
	return cmd
}

package uphost

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `uphost` root command (list-only).
// Mirrors cmd/uphost.go NewCmdUPHost + NewCmdUPHostList.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uphost",
		Short: "List UPHost instances",
		Long:  `List UPHost instances`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))

	return cmd
}

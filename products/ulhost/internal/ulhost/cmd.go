package ulhost

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `ulhost` root command.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ulhost",
		Short: "List,create,delete,restart or resize ULHost instance",
		Long:  `List,create,delete,restart or resize ULHost instance`,
		Args:  cobra.NoArgs,
	}

	return cmd
}

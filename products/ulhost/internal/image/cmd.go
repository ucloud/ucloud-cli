package image

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `image` root command for ulhost, mirroring the uhost
// image subcommand pattern.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "List ULHost images",
		Long:  `List available images for ULHost instances`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))

	return cmd
}

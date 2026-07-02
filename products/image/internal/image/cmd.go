package image

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `image` root command and mounts the 4 subcommands.
// Mirrors cmd/image.go NewCmdUImage (same AddCommand order: list, copy, delete,
// create). The create subcommand is image's OWN copy of uhost's create-image
// (newCreateImage) — image no longer borrows NewCmdUhostCreateImage from cmd.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "List and manipulate images",
		Long:  `List and manipulate images`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCopy(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newCreateImage(ctx))

	return cmd
}

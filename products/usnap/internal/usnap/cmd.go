package usnap

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `usnap` root command and mounts the subcommands.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usnap",
		Short: "Manage USnap (UCloud Disk Snapshot Service)",
		Long:  "Manage USnap (UCloud Disk Snapshot Service)",
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newDelete(ctx))
	return cmd
}

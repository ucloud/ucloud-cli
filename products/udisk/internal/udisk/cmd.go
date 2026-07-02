package udisk

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `udisk` root command and mounts the 11 subcommands.
// Mirrors cmd/disk.go NewCmdDisk (same AddCommand order).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "udisk",
		Short: "Read and manipulate udisk instances",
		Long:  "Read and manipulate udisk instances",
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newAttach(ctx))
	cmd.AddCommand(newDetach(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newClone(ctx))
	cmd.AddCommand(newExpand(ctx))
	cmd.AddCommand(newSnapshot(ctx))
	cmd.AddCommand(newRestore(ctx))
	cmd.AddCommand(newSnapshotList(ctx))
	cmd.AddCommand(newSnapshotDelete(ctx))
	return cmd
}

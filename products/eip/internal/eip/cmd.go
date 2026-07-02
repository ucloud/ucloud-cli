package eip

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `eip` root command and mounts the 9 subcommands.
// Mirrors cmd/eip.go NewCmdEIP (same AddCommand order).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eip",
		Short: "List,allocate and release EIP",
		Long:  `Manipulate EIP, such as list,allocate and release`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newAllocate(ctx))
	cmd.AddCommand(newRelease(ctx))
	cmd.AddCommand(newBind(ctx))
	cmd.AddCommand(newUnbind(ctx))
	cmd.AddCommand(newModifyBandwidth(ctx))
	cmd.AddCommand(newSetChargeMode(ctx))
	cmd.AddCommand(newJoinSharedBW(ctx))
	cmd.AddCommand(newLeaveSharedBW(ctx))
	return cmd
}

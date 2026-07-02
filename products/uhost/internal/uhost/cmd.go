package uhost

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `uhost` root command and mounts the 14 subcommands in
// the same AddCommand order as cmd/uhost.go NewCmdUHost: list, create, delete,
// stop, start, restart, poweroff, resize, clone, reset-password, reinstall-os,
// create-image, isolation-group (subtree), leave-isolation-group.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uhost",
		Short: "List,create,delete,stop,restart,poweroff or resize UHost instance",
		Long:  `List,create,delete,stop,restart,poweroff or resize UHost instance`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newStop(ctx))
	cmd.AddCommand(newStart(ctx))
	cmd.AddCommand(newReboot(ctx))
	cmd.AddCommand(newPoweroff(ctx))
	cmd.AddCommand(newResize(ctx))
	cmd.AddCommand(newClone(ctx))
	cmd.AddCommand(newResetPassword(ctx))
	cmd.AddCommand(newReinstallOS(ctx))
	cmd.AddCommand(newCreateImage(ctx))
	cmd.AddCommand(newIsolationGroup(ctx))
	cmd.AddCommand(newLeaveIsolationGroup(ctx))

	return cmd
}

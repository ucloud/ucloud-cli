package ulhost

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `ulhost` root command and mounts the subcommands in
// the same AddCommand order as the uhost product: list, create, delete, stop,
// start, restart, poweroff, reset-password, reinstall-os, modify-attribute,
// bundles, price.
//
// NOTE: The backend API also exposes UpdateULHostInstanceFirewall,
// ModifyULHostProxyIp, CheckULHostResourceCapacity, and share-bandwidth
// management, but the public ucompshare SDK does not yet support these
// operations. When the SDK adds them, corresponding CLI commands should be
// added here following the same pattern.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ulhost",
		Short: "List,create,delete,stop,restart,poweroff or resize ULHost instance",
		Long:  `List,create,delete,stop,restart,poweroff or resize ULHost instance`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newStop(ctx))
	cmd.AddCommand(newStart(ctx))
	cmd.AddCommand(newReboot(ctx))
	cmd.AddCommand(newPoweroff(ctx))
	cmd.AddCommand(newResetPassword(ctx))
	cmd.AddCommand(newReinstallOS(ctx))
	cmd.AddCommand(newModifyAttribute(ctx))
	cmd.AddCommand(newBundles(ctx))
	cmd.AddCommand(newPrice(ctx))

	return cmd
}

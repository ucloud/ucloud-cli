package firewall

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `firewall` root command and mounts the 9 subcommands.
// Mirrors cmd/firewall.go NewCmdFirewall (same AddCommand order).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firewall",
		Short: "List and manipulate extranet firewall",
		Long:  `List and manipulate extranet firewall`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newAddRule(ctx))
	cmd.AddCommand(newDeleteRule(ctx))
	cmd.AddCommand(newApply(ctx))
	cmd.AddCommand(newCopy(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newResource(ctx))
	cmd.AddCommand(newUpdate(ctx))

	return cmd
}

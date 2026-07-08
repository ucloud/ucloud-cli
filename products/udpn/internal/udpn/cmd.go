package udpn

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `udpn` root command.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "udpn",
		Short: "List and manipulate udpn instances",
		Long:  "List and manipulate udpn instances",
	}

	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newModifyBW(ctx))

	return cmd
}

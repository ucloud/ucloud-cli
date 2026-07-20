package ugn

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newNetwork ucloud ugn network
func newNetwork(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Manage ugn network instances",
		Long:  "Manage ugn network instances",
	}

	cmd.AddCommand(newNetworkAttach(ctx))
	cmd.AddCommand(newNetworkDetach(ctx))

	return cmd
}

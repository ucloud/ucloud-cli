package ext

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUHost builds `ucloud ext uhost`.
func newUHost(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uhost",
		Short: "extended uhost commands",
		Long:  "extended uhost commands",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newUHostSwitchEIP(ctx))
	return cmd
}

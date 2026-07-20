package ugn

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newBW ucloud ugn bw
func newBW(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bw",
		Short: "List and manipulate ugn bandwidth packages",
		Long:  "List and manipulate ugn bandwidth packages",
	}

	cmd.AddCommand(newBWCreate(ctx))
	cmd.AddCommand(newBWDelete(ctx))
	cmd.AddCommand(newBWList(ctx))
	cmd.AddCommand(newBWModifyBandwidth(ctx))

	return cmd
}

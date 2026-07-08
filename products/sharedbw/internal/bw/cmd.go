package bw

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand returns the ucloud bw command tree.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bw",
		Short: "Manipulate bandwidth package and shared bandwidth",
		Long:  "Manipulate bandwidth package and shared bandwidth",
	}
	cmd.AddCommand(newPkg(ctx))
	cmd.AddCommand(newShared(ctx))
	return cmd
}

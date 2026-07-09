package bw

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newShared returns ucloud bw shared.
func newShared(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shared",
		Short: "Create and manipulate shared bandwidth instances",
		Long:  "Create and manipulate shared bandwidth instances",
	}
	cmd.AddCommand(newSharedCreate(ctx))
	cmd.AddCommand(newSharedList(ctx))
	cmd.AddCommand(newSharedResize(ctx))
	cmd.AddCommand(newSharedDelete(ctx))
	return cmd
}

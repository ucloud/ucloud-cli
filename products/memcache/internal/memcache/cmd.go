package memcache

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand returns the ucloud memcache command tree.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memcache",
		Short: "List and manipulate memcache instances",
		Long:  "List and manipulate memcache instances",
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newRestart(ctx))
	return cmd
}

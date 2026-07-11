package clickhouse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `uclickhouse` root command and mounts the subcommands.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uclickhouse",
		Short: "Manage UClickhouse clusters",
		Long:  "Manage UClickhouse clusters",
		Args:  noArgs,
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newExpand(ctx))
	cmd.AddCommand(newResize(ctx))
	cmd.AddCommand(newRestart(ctx))
	cmd.AddCommand(newCreateOption(ctx))
	return cmd
}

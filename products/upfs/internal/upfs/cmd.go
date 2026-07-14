package upfs

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `upfs` root command and mounts the subcommands.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upfs",
		Short: "Manage UPFS (UCloud Parallel File Storage) volumes",
		Long:  "Manage UPFS (UCloud Parallel File Storage) volumes",
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newDelete(ctx))
	return cmd
}

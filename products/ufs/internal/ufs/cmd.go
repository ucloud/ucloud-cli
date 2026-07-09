package ufs

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `ufs` root command and mounts the subcommands.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ufs",
		Short: "Manage UFS (UCloud File Storage) volumes",
		Long:  "Manage UFS (UCloud File Storage) volumes",
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newDelete(ctx))
	return cmd
}

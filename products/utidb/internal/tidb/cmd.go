package tidb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `utidb` root command and mounts all verbs.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "utidb",
		Short: "Manipulate UTiDB instances on UCloud platform",
		Long:  helpUTiDBRoot,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newBackup(ctx))
	cmd.AddCommand(newListBackup(ctx))
	cmd.AddCommand(newScaleNode(ctx))
	cmd.AddCommand(newResizeDisk(ctx))
	cmd.AddCommand(newModifySpec(ctx))
	cmd.AddCommand(newListSpecs(ctx))
	return cmd
}

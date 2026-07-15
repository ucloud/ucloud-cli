package uhadoop

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the top-level `uhadoop` command.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uhadoop",
		Short: "List,create,delete,describe UHadoop clusters and manage nodes and services",
		Long:  `List,create,delete,describe UHadoop clusters and manage nodes and services`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newAddNode(ctx))
	cmd.AddCommand(newListNodeType(ctx))
	cmd.AddCommand(newListFrameworkApp(ctx))
	cmd.AddCommand(newRestartService(ctx))
	cmd.AddCommand(newUpgradeNode(ctx))
	cmd.AddCommand(newUpgradeDisk(ctx))
	return cmd
}

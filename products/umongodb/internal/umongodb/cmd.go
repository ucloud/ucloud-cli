package umongodb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand assembles the umongodb command tree.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   productName,
		Short: "Manipulate MongoDB on UCloud platform",
		Long:  "Manipulate MongoDB on UCloud platform",
	}

	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newListVersions(ctx))
	cmd.AddCommand(newListTemplates(ctx))
	cmd.AddCommand(newListMachineSpecs(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newCreateReplset(ctx))
	cmd.AddCommand(newCreateSharded(ctx))
	cmd.AddCommand(newStart(ctx))
	cmd.AddCommand(newStop(ctx))
	cmd.AddCommand(newRestart(ctx))
	cmd.AddCommand(newDeleteReplset(ctx))
	cmd.AddCommand(newDeleteSharded(ctx))

	return cmd
}

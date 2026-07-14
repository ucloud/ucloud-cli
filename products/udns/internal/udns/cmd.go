package udns

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func NewCommand(ctx *cli.Context) *cobra.Command {
	root := &cobra.Command{
		Use:   "udns",
		Short: "List and manipulate ucloud private dns(udns) instance and record",
		Long:  "List and manipulate ucloud private dns(udns) instance and record",
	}
	root.AddCommand(newCreateCommand(ctx))
	root.AddCommand(newListCommand(ctx))
	root.AddCommand(newModifyCommand(ctx))
	root.AddCommand(newAssociateVPCCommand(ctx))
	root.AddCommand(newDisassociateVPCCommand(ctx))
	root.AddCommand(newRecordCommand(ctx))
	return root
}

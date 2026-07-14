package udns

import (
	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

type udns struct{}

func New() cli.Product {
	return udns{}
}
func (u udns) Metadata() cli.Metadata {
	return cli.Metadata{
		Name:     "udns",
		Commands: []string{"udns"},
	}
}

func (u udns) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{
		newUdnsCommand(ctx),
	}
}

func newUdnsCommand(ctx *cli.Context) *cobra.Command {
	root := &cobra.Command{
		Use:   "udns",
		Short: "List and manipulate ucloud private dns(udns) instance and record",
		Long:  "List and manipulate ucloud private dns(udns) instance and record",
	}
	root.AddCommand(NewCreateCommand(ctx))
	root.AddCommand(newListCommand(ctx))
	root.AddCommand(newModifyCommand(ctx))
	root.AddCommand(newAssociateVPCCommand(ctx))
	root.AddCommand(newDisassociateVPCCommand(ctx))
	root.AddCommand(newRecordCommand(ctx))
	return root
}

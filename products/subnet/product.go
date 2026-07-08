package subnet

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalsubnet "github.com/ucloud/ucloud-cli/products/subnet/internal/subnet"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "subnet", Commands: []string{"subnet"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalsubnet.NewCommand(ctx)}
}

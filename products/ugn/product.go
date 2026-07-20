package ugn

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalugn "github.com/ucloud/ucloud-cli/products/ugn/internal/ugn"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "ugn", Commands: []string{"ugn"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalugn.NewCommand(ctx)}
}

var _ cli.Product = (*product)(nil)

package uhadoop

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaluhadoop "github.com/ucloud/ucloud-cli/products/uhadoop/internal/uhadoop"
)

type product struct{}

// New returns the uhadoop product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "uhadoop", Commands: []string{"uhadoop"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internaluhadoop.NewCommand(ctx)}
}

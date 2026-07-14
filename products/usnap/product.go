package usnap

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalusnap "github.com/ucloud/ucloud-cli/products/usnap/internal/usnap"
)

type product struct{}

// New returns the usnap product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "usnap", Commands: []string{"usnap"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalusnap.NewCommand(ctx)}
}

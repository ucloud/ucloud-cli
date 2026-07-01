package udisk

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaludisk "github.com/ucloud/ucloud-cli/products/udisk/internal/udisk"
)

type product struct{}

// New returns the udisk product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "udisk", Commands: []string{"udisk"}}
}

func (product) NewCommand(ctx *cli.Context) *cobra.Command { return internaludisk.NewCommand(ctx) }

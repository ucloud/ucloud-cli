package eip

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaleip "github.com/ucloud/ucloud-cli/products/eip/internal/eip"
)

type product struct{}

// New returns the eip product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "eip", Commands: []string{"eip"}}
}

func (product) NewCommand(ctx *cli.Context) *cobra.Command { return internaleip.NewCommand(ctx) }

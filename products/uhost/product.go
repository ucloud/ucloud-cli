package uhost

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaluhost "github.com/ucloud/ucloud-cli/products/uhost/internal/uhost"
)

type product struct{}

// New returns the uhost product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "uhost", Commands: []string{"uhost"}}
}

func (product) NewCommand(ctx *cli.Context) *cobra.Command { return internaluhost.NewCommand(ctx) }

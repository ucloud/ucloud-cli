package uphost

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaluphost "github.com/ucloud/ucloud-cli/products/uphost/internal/uphost"
)

type product struct{}

// New returns the uphost product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "uphost", Commands: []string{"uphost"}}
}

func (product) NewCommand(ctx *cli.Context) *cobra.Command { return internaluphost.NewCommand(ctx) }

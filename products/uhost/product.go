package uhost

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalimage "github.com/ucloud/ucloud-cli/products/uhost/internal/image"
	internaluhost "github.com/ucloud/ucloud-cli/products/uhost/internal/uhost"
)

type product struct{}

// New returns the uhost product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "uhost", Commands: []string{"uhost", "image"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{
		internaluhost.NewCommand(ctx),
		internalimage.NewCommand(ctx),
	}
}

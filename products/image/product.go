package image

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalimage "github.com/ucloud/ucloud-cli/products/image/internal/image"
)

type product struct{}

// New returns the image product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "image", Commands: []string{"image"}}
}

func (product) NewCommand(ctx *cli.Context) *cobra.Command { return internalimage.NewCommand(ctx) }

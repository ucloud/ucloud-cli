package css

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalcss "github.com/ucloud/ucloud-cli/products/css/internal/css"
)

type product struct{}

// New returns the css product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "css", Commands: []string{"css"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalcss.NewCommand(ctx)}
}

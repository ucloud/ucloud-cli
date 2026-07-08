package ulb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalulb "github.com/ucloud/ucloud-cli/products/ulb/internal/ulb"
)

type product struct{}

// New returns the ulb product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "ulb", Commands: []string{"ulb"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalulb.NewCommand(ctx)}
}

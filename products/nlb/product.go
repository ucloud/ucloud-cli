package nlb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalnlb "github.com/ucloud/ucloud-cli/products/nlb/internal/nlb"
)

type product struct{}

// New returns the nlb product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "nlb", Commands: []string{"nlb"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalnlb.NewCommand(ctx)}
}

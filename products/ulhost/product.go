package ulhost

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalulhost "github.com/ucloud/ucloud-cli/products/ulhost/internal/ulhost"
)

type product struct{}

// New returns the ulhost product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "ulhost", Commands: []string{"ulhost"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{
		internalulhost.NewCommand(ctx),
	}
}

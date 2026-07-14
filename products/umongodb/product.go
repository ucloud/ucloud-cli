package umongodb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalumongodb "github.com/ucloud/ucloud-cli/products/umongodb/internal/umongodb"
)

type product struct{}

// New returns the umongodb product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "umongodb", Commands: []string{"umongodb"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalumongodb.NewCommand(ctx)}
}

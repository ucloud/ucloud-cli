package ukafka

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalukafka "github.com/ucloud/ucloud-cli/products/ukafka/internal/ukafka"
)

type product struct{}

// New returns the ukafka product (registered via hack/gen-products)
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "ukafka", Commands: []string{"ukafka"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalukafka.NewCommand(ctx)}
}

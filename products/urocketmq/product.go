package urocketmq

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internal "github.com/ucloud/ucloud-cli/products/urocketmq/internal"
)

type product struct{}

// New returns the urocketmq product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "urocketmq", Commands: []string{"urocketmq"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internal.NewCommand(ctx)}
}

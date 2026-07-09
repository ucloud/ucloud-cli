package udpn

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaludpn "github.com/ucloud/ucloud-cli/products/udpn/internal/udpn"
)

type product struct{}

// New returns the udpn product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "udpn", Commands: []string{"udpn"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internaludpn.NewCommand(ctx)}
}

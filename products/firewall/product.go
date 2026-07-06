package firewall

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalfirewall "github.com/ucloud/ucloud-cli/products/firewall/internal/firewall"
)

type product struct{}

// New returns the firewall product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "firewall", Commands: []string{"firewall"}}
}

func (product) NewCommand(ctx *cli.Context) *cobra.Command { return internalfirewall.NewCommand(ctx) }

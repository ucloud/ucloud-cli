package unet

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaleip "github.com/ucloud/ucloud-cli/products/unet/internal/eip"
	internalfirewall "github.com/ucloud/ucloud-cli/products/unet/internal/firewall"
)

type product struct{}

// New returns the unet product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "unet", Commands: []string{"eip", "firewall"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{
		internaleip.NewCommand(ctx),
		internalfirewall.NewCommand(ctx),
	}
}

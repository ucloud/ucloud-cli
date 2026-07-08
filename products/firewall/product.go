package firewall

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalfirewall "github.com/ucloud/ucloud-cli/products/firewall/internal/firewall"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "firewall", Commands: []string{"firewall"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalfirewall.NewCommand(ctx)}
}

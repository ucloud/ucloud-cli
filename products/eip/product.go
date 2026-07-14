package eip

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaleip "github.com/ucloud/ucloud-cli/products/eip/internal/eip"
	internalext "github.com/ucloud/ucloud-cli/products/eip/internal/ext"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "eip", Commands: []string{"eip", "ext"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internaleip.NewCommand(ctx), internalext.NewCommand(ctx)}
}

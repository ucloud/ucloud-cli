package sharedbw

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalbw "github.com/ucloud/ucloud-cli/products/sharedbw/internal/bw"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "sharedbw", Commands: []string{"bw"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalbw.NewCommand(ctx)}
}

package vpc

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalvpc "github.com/ucloud/ucloud-cli/products/vpc/internal/vpc"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "vpc", Commands: []string{"vpc"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalvpc.NewCommand(ctx)}
}

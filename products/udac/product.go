package udac

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaludac "github.com/ucloud/ucloud-cli/products/udac/internal/udac"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "udac", Commands: []string{"udac"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internaludac.NewCommand(ctx)}
}
